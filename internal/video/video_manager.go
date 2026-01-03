package video

import (
	"fmt"
	"os"
	"os/exec"
)

// VideoOptions contains configurable options for video playback
type VideoOptions struct {
	InputFormat string
	VideoSize   string
	WindowTitle string
	FFlags      string
	Flags       string
	VF          string
	AlwaysOnTop bool
}

// DefaultVideoOptions returns the default video options
func DefaultVideoOptions() VideoOptions {
	return VideoOptions{
		InputFormat: "mjpeg",
		VideoSize:   "1920x1080",
		WindowTitle: "Capture Clam",
		FFlags:      "nobuffer",
		Flags:       "low_delay",
		VF:          "setpts=0",
		AlwaysOnTop: true,
	}
}

// VideoManager handles video preview using ffplay
type VideoManager interface {
	Start() error
	Wait() error
	Stop() error
}

// videoManager implements VideoManager
type videoManager struct {
	device string
	cmd    *exec.Cmd
	opts   VideoOptions
}

// NewVideoManager creates a new video manager
func NewVideoManager(device string) VideoManager {
	return NewVideoManagerWithOptions(device, DefaultVideoOptions())
}

// NewVideoManagerWithOptions creates a new video manager with custom options
func NewVideoManagerWithOptions(device string, opts VideoOptions) VideoManager {
	return &videoManager{
		device: device,
		opts:   opts,
	}
}

// Start launches the video preview window
func (vm *videoManager) Start() error {
	fmt.Println("📺 Launching video preview window...")
	args := []string{
		"-f", "v4l2",
		"-input_format", vm.opts.InputFormat,
		"-video_size", vm.opts.VideoSize,
		"-i", vm.device,
		"-an", // No audio in ffplay (handled by pactl)
		"-fflags", vm.opts.FFlags,
		"-flags", vm.opts.Flags,
		"-framedrop",
		"-window_title", vm.opts.WindowTitle,
		"-vf", vm.opts.VF,
	}

	if vm.opts.AlwaysOnTop {
		args = append(args, "-alwaysontop")
	}

	vm.cmd = exec.Command("ffplay", args...)

	// Set PipeWire latency for the video process
	vm.cmd.Env = append(os.Environ(), "PIPEWIRE_LATENCY=64/48000")

	// Start video
	if err := vm.cmd.Start(); err != nil {
		return fmt.Errorf("error starting video: %w", err)
	}

	return nil
}

// Wait waits for the video process to finish
func (vm *videoManager) Wait() error {
	if vm.cmd == nil {
		return nil
	}
	return vm.cmd.Wait()
}

// Stop kills the video process
func (vm *videoManager) Stop() error {
	if vm.cmd == nil || vm.cmd.Process == nil {
		return nil
	}
	return vm.cmd.Process.Kill()
}
