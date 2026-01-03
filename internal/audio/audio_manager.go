package audio

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// AudioManager handles audio loopback module management
type AudioManager interface {
	Start() error
	Stop() error
	GetLoopbackID() string
	Cleanup() error
}

// audioManager implements AudioManager
type audioManager struct {
	loopbackID string
	source     string
}

// NewAudioManager creates a new audio manager
func NewAudioManager(source string) AudioManager {
	return &audioManager{
		source: source,
	}
}

// Start loads the audio loopback module
func (am *audioManager) Start() error {
	cmd := exec.Command("pactl", "load-module", "module-loopback", "latency_msec=1", "source="+am.source)

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error loading audio module: %w", err)
	}

	am.loopbackID = strings.TrimSpace(out.String())
	return nil
}

// Stop unloads the audio loopback module
func (am *audioManager) Stop() error {
	if am.loopbackID == "" {
		return nil
	}

	// Use context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pactl", "unload-module", am.loopbackID)
	return cmd.Run()
}

// GetLoopbackID returns the pulseaudio loopback module ID
func (am *audioManager) GetLoopbackID() string {
	return am.loopbackID
}

// Cleanup unloads all PulseAudio loopback modules
func (am *audioManager) Cleanup() error {
	// Get list of modules
	cmd := exec.Command("pactl", "list", "modules", "short")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error listing modules: %w", err)
	}

	// Parse output to find loopback modules
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

	// Unload each module with timeout
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
