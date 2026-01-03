package capture

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
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

// Service handles video and audio capture operations.
type Service struct {
	videoDevice string
	audioDevice string
	cmd         *exec.Cmd
}

// New creates a new capture service.
func New(videoDevice, audioDevice string) *Service {
	return &Service{
		videoDevice: videoDevice,
		audioDevice: audioDevice,
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
	gstArgs := []string{"-v",
		"v4l2src", "device=" + s.videoDevice, "!",
		"image/jpeg,framerate=" + "60/1", "!", "jpegdec", "!", "videoconvert", "!", "queue", "leaky=downstream", "max-size-buffers=2", "!", "autovideosink", "sync=false",
	}

	if s.audioDevice != "" {
		gstArgs = append(gstArgs,
			"pulsesrc", "device="+s.audioDevice, "!",
			"audio/x-raw,format=S16LE,rate=48000,channels=2", "!", "audioconvert", "!", "audioresample", "!", "queue", "leaky=downstream", "max-size-buffers=32", "!", "autoaudiosink", "sync=false",
		)
	}

	s.cmd = exec.Command("gst-launch-1.0", gstArgs...)
	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stderr
	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("error starting gstreamer pipeline: %w", err)
	}
	return nil
}

// Wait blocks until the capture process finishes.
func (s *Service) Wait() error {
	if s.cmd == nil {
		return nil
	}
	return s.cmd.Wait()
}

// Stop terminates the capture process.
func (s *Service) Stop() error {
	if s.cmd == nil || s.cmd.Process == nil {
		return nil
	}
	if err := s.cmd.Process.Kill(); err != nil {
		if errors.Is(err, os.ErrProcessDone) || strings.Contains(err.Error(), "already finished") {
			return nil
		}
		return err
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
