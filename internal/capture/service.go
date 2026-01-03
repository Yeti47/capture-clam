package capture

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-gst/go-gst/gst"
)

// AudioDevice represents an audio source device
type AudioDevice struct {
	Name        string
	Description string
}

// VideoDevice represents a video device
type VideoDevice struct {
	Device string
	Name   string
}

// Config contains configurable parameters for the capture pipeline.
type Config struct {
	Framerate      string // e.g., "60/1"
	VideoFormat    string // e.g., "image/jpeg"
	AudioFormat    string // e.g., "S16LE"
	AudioRate      int    // e.g., 48000
	AudioChannels  int    // e.g., 2
	VideoQueueSize int    // max-size-buffers for video queue
	AudioQueueSize int    // max-size-buffers for audio queue
}

// DefaultConfig returns sensible defaults for capture.
func DefaultConfig() Config {
	return Config{
		Framerate:      "60/1",
		VideoFormat:    "image/jpeg",
		AudioFormat:    "S16LE",
		AudioRate:      48000,
		AudioChannels:  2,
		VideoQueueSize: 2,
		AudioQueueSize: 32,
	}
}

// Service handles video and audio capture operations.
type Service struct {
	videoDevice string
	audioDevice string
	config      Config
	pipeline    *gst.Pipeline
	done        chan bool
}

// New creates a new capture service with default configuration.
func New(videoDevice, audioDevice string) *Service {
	return NewWithConfig(videoDevice, audioDevice, DefaultConfig())
}

// NewWithConfig creates a new capture service with custom configuration.
func NewWithConfig(videoDevice, audioDevice string, config Config) *Service {
	return &Service{
		videoDevice: videoDevice,
		audioDevice: audioDevice,
		config:      config,
	}
}

// ListAudioDevices returns available audio source devices.
func ListAudioDevices() ([]AudioDevice, error) {
	cmd := exec.Command("pactl", "list", "sources", "short")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to list audio devices: %w", err)
	}

	var devices []AudioDevice
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			devices = append(devices, AudioDevice{
				Name:        parts[1],
				Description: strings.Join(parts[2:], " "),
			})
		}
	}
	return devices, nil
}

// ListVideoDevices returns available video devices.
func ListVideoDevices() ([]VideoDevice, error) {
	cmd := exec.Command("v4l2-ctl", "--list-devices")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to list video devices: %w", err)
	}

	var devices []VideoDevice
	output := out.String()
	lines := strings.Split(output, "\n")

	var currentName string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "/dev/video") {
			devices = append(devices, VideoDevice{
				Device: line,
				Name:   currentName,
			})
		} else {
			currentName = line
		}
	}

	return devices, nil
}

// GetDefaultVideoDevice returns the first available video device.
func GetDefaultVideoDevice() (string, error) {
	devices, err := ListVideoDevices()
	if err != nil {
		return "", err
	}
	if len(devices) == 0 {
		return "", fmt.Errorf("no video devices found")
	}
	return devices[0].Device, nil
}

// Start launches the gstreamer preview with audio and video.
func (s *Service) Start() error {
	// Initialize GStreamer
	gst.Init(nil)

	// Build pipeline description
	pipelineDesc := fmt.Sprintf(
		"v4l2src device=%s ! %s,framerate=%s ! jpegdec ! videoconvert ! queue leaky=downstream max-size-buffers=%d name=vqueue ! xvimagesink name=videosink sync=false",
		s.videoDevice,
		s.config.VideoFormat,
		s.config.Framerate,
		s.config.VideoQueueSize,
	)

	if s.audioDevice != "" {
		pipelineDesc += fmt.Sprintf(
			" pulsesrc device=%s ! audio/x-raw,format=%s,rate=%d,channels=%d ! audioconvert ! audioresample ! queue leaky=downstream max-size-buffers=%d name=aqueue ! autoaudiosink sync=false",
			s.audioDevice,
			s.config.AudioFormat,
			s.config.AudioRate,
			s.config.AudioChannels,
			s.config.AudioQueueSize,
		)
	}

	// Create pipeline from description
	pipeline, err := gst.NewPipelineFromString(pipelineDesc)
	if err != nil {
		return fmt.Errorf("failed to create pipeline: %w", err)
	}
	s.pipeline = pipeline
	s.done = make(chan bool)

	// Handle messages on the bus
	go func() {
		bus := s.pipeline.GetBus()
		for {
			msg := bus.TimedPopFiltered(gst.ClockTimeNone, gst.MessageEOS|gst.MessageError)
			if msg == nil {
				break
			}

			switch msg.Type() {
			case gst.MessageEOS:
				s.done <- true
				return
			case gst.MessageError:
				gerr := msg.ParseError()
				fmt.Fprintf(os.Stderr, "Error from %s: %s\n", msg.Source(), gerr.Error())
				s.done <- true
				return
			}
		}
	}()

	// Set pipeline to playing
	if err := s.pipeline.SetState(gst.StatePlaying); err != nil {
		return fmt.Errorf("failed to set pipeline to playing: %w", err)
	}

	return nil
}

// Wait blocks until the capture process finishes.
func (s *Service) Wait() error {
	if s.done == nil {
		return nil
	}
	<-s.done
	return nil
}

// Stop terminates the capture process.
func (s *Service) Stop() error {
	if s.pipeline != nil {
		s.pipeline.SetState(gst.StateNull)
	}
	if s.done != nil {
		select {
		case s.done <- true:
		default:
		}
	}
	return nil
}

// CleanupLoopbackModules removes any PulseAudio loopback modules left from interrupted runs.
func CleanupLoopbackModules() error {
	cmd := exec.Command("pactl", "list", "modules", "short")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error listing modules: %w", err)
	}

	var loopbackIDs []string
	scanner := bufio.NewScanner(strings.NewReader(out.String()))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == "module-loopback" {
			loopbackIDs = append(loopbackIDs, parts[0])
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error scanning modules output: %w", err)
	}

	for _, id := range loopbackIDs {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		cmd := exec.CommandContext(ctx, "pactl", "unload-module", id)
		if err := cmd.Run(); err != nil {
			cancel()
			return fmt.Errorf("error unloading module %s: %w", id, err)
		}
		cancel()
	}

	return nil
}
