package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Yeti47/capture-clam/internal/capture"

	"github.com/spf13/cobra"
)

var (
	audioSource         string
	videoSource         string
	resolvedAudioSource string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the video capture application",
	Long:  `Start the ultra-low latency audio and video preview using gstreamer.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚀 Initializing Ultra-Low Latency System...")

		// Validate and resolve audio source
		if audioSource == "" {
			fmt.Println("❌ Audio source is required. Use --audio-source or list available sources with 'list-audio'")
			os.Exit(1)
		}

		// Try to parse as index first
		if index, err := strconv.Atoi(audioSource); err == nil {
			// It's a number, get device by index
			audioDevices, err := capture.ListAudioDevices()
			if err != nil {
				fmt.Printf("❌ Error listing audio devices: %v\n", err)
				os.Exit(1)
			}

			if index < 1 || index > len(audioDevices) {
				fmt.Printf("❌ Invalid audio source index: %d. Valid range is 1-%d\n", index, len(audioDevices))
				os.Exit(1)
			}

			resolvedAudioSource = audioDevices[index-1].Name // Convert to 0-based index
		} else {
			// Not a number, use the string directly as device name
			resolvedAudioSource = audioSource
		}

		fmt.Printf("🎤 Using audio device: %s\n", resolvedAudioSource)

		// Get video source
		usingDefaultVideo := videoSource == ""
		if videoSource == "" {
			defaultVideo, err := capture.GetDefaultVideoDevice()
			if err != nil {
				fmt.Printf("❌ Error getting default video device: %v\n", err)
				os.Exit(1)
			}
			videoSource = defaultVideo
		} else {
			// Try to parse video source as index
			if index, err := strconv.Atoi(videoSource); err == nil {
				// It's a number, get device by index
				videoDevices, err := capture.ListVideoDevices()
				if err != nil {
					fmt.Printf("❌ Error listing video devices: %v\n", err)
					os.Exit(1)
				}

				if index < 1 || index > len(videoDevices) {
					fmt.Printf("❌ Invalid video source index: %d. Valid range is 1-%d\n", index, len(videoDevices))
					os.Exit(1)
				}

				videoSource = videoDevices[index-1].Device // Convert to 0-based index
			}
			// If not a number, use the string directly as device name
		}

		if usingDefaultVideo {
			fmt.Printf("📹 Using default video device: %s\n", videoSource)
		} else {
			fmt.Printf("📹 Using video device: %s\n", videoSource)
		}

		// Create capture service with selected devices
		svc := capture.New(videoSource, resolvedAudioSource)

		fmt.Println("📺 Launching video preview (gstreamer)...")

		// Setup cleanup handler
		cleanupDone := make(chan bool)
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-sigChan
			fmt.Println("🧹 Shutting down...")
			if err := svc.Stop(); err != nil {
				fmt.Printf("❌ Error stopping capture: %v\n", err)
			}
			cleanupDone <- true
		}()

		// Start capture
		if err := svc.Start(); err != nil {
			fmt.Printf("❌ %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✨ Running! Press Ctrl+C in this terminal to exit.")

		// Wait for capture to finish or signal
		go func() {
			svc.Wait()
			sigChan <- syscall.SIGTERM
		}()

		<-cleanupDone
		fmt.Println("👋 Goodbye!")
	},
}

func init() {
	runCmd.Flags().StringVar(&audioSource, "audio-source", "", "Audio source device (index from list-audio or device name)")
	runCmd.Flags().StringVar(&videoSource, "video-source", "", "Video device (index from list-video or device name, optional, defaults to first available)")
	runCmd.MarkFlagRequired("audio-source")
}
