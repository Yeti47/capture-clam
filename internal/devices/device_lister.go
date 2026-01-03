package devices

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
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

// DeviceLister handles listing available audio and video devices
type DeviceLister interface {
	ListAudioDevices() ([]AudioDevice, error)
	ListVideoDevices() ([]VideoDevice, error)
	GetDefaultVideoDevice() (string, error)
}

// deviceLister implements DeviceLister
type deviceLister struct{}

// NewDeviceLister creates a new device lister
func NewDeviceLister() DeviceLister {
	return &deviceLister{}
}

// ListAudioDevices returns a list of available audio sources
func (dl *deviceLister) ListAudioDevices() ([]AudioDevice, error) {
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

// ListVideoDevices returns a list of available video devices
func (dl *deviceLister) ListVideoDevices() ([]VideoDevice, error) {
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

		// Check if line starts with /dev/video (device path)
		if strings.HasPrefix(line, "/dev/video") {
			// This is a device path
			devices = append(devices, VideoDevice{
				Device: line,
				Name:   currentName,
			})
		} else {
			// This is a device name
			currentName = line
		}
	}

	return devices, nil
}

// GetDefaultVideoDevice returns the first available video device
func (dl *deviceLister) GetDefaultVideoDevice() (string, error) {
	devices, err := dl.ListVideoDevices()
	if err != nil {
		return "", err
	}
	if len(devices) == 0 {
		return "", fmt.Errorf("no video devices found")
	}
	return devices[0].Device, nil
}
