package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Yeti47/capture-clam/internal/audio"
	"github.com/Yeti47/capture-clam/internal/devices"
	"github.com/Yeti47/capture-clam/internal/video"

	"github.com/spf13/cobra"
)

var (
	audioSource         string
	videoSource         string
	videoSize           string
	inputFormat         string
	windowTitle         string
	disableAlwaysOnTop  bool
	resolvedAudioSource string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the video capture application",
	Long:  `Start the ultra-low latency video capture with audio loopback.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚀 Initializing Ultra-Low Latency System...")

		// Initialize device lister
		deviceLister := devices.NewDeviceLister()

		// Validate and resolve audio source
		if audioSource == "" {
			fmt.Println("❌ Audio source is required. Use --audio-source or list available sources with 'list-audio'")
			os.Exit(1)
		}

		// Try to parse as index first
		if index, err := strconv.Atoi(audioSource); err == nil {
			// It's a number, get device by index
			audioDevices, err := deviceLister.ListAudioDevices()
			if err != nil {
				fmt.Printf("❌ Error listing audio devices: %v\n", err)
				os.Exit(1)
			}

			if index < 1 || index > len(audioDevices) {
				fmt.Printf("❌ Invalid audio source index: %d. Valid range is 1-%d\n", index, len(audioDevices))
				os.Exit(1)
			}

			resolvedAudioSource = audioDevices[index-1].Name // Convert to 0-based index
			fmt.Printf("📹 Using audio device: %s\n", resolvedAudioSource)
		} else {
			// Not a number, use the string directly as device name
			resolvedAudioSource = audioSource
		}

		// Get video source
		if videoSource == "" {
			defaultVideo, err := deviceLister.GetDefaultVideoDevice()
			if err != nil {
				fmt.Printf("❌ Error getting default video device: %v\n", err)
				os.Exit(1)
			}
			videoSource = defaultVideo
			fmt.Printf("📹 Using default video device: %s\n", videoSource)
		} else {
			// Try to parse video source as index
			if index, err := strconv.Atoi(videoSource); err == nil {
				// It's a number, get device by index
				videoDevices, err := deviceLister.ListVideoDevices()
				if err != nil {
					fmt.Printf("❌ Error listing video devices: %v\n", err)
					os.Exit(1)
				}

				if index < 1 || index > len(videoDevices) {
					fmt.Printf("❌ Invalid video source index: %d. Valid range is 1-%d\n", index, len(videoDevices))
					os.Exit(1)
				}

				videoSource = videoDevices[index-1].Device // Convert to 0-based index
				fmt.Printf("📹 Using video device: %s\n", videoSource)
			}
			// If not a number, use the string directly as device name
		}

		// Initialize managers
		audioMgr := audio.NewAudioManager(resolvedAudioSource)

		// Create video options
		videoOpts := video.DefaultVideoOptions()
		if videoSize != "" {
			videoOpts.VideoSize = videoSize
		}
		if inputFormat != "" {
			videoOpts.InputFormat = inputFormat
		}
		if windowTitle != "" {
			videoOpts.WindowTitle = windowTitle
		}
		videoOpts.AlwaysOnTop = !disableAlwaysOnTop

		videoMgr := video.NewVideoManagerWithOptions(videoSource, videoOpts)

		// Start audio
		fmt.Printf("🎤 Linking audio source: %s\n", resolvedAudioSource)
		if err := audioMgr.Start(); err != nil {
			fmt.Printf("❌ %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Audio linked (Module ID: %s)\n", audioMgr.GetLoopbackID())

		// Setup cleanup handler
		cleanupDone := make(chan bool)
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-sigChan
			fmt.Println("🧹 Shutting down and unloading audio...")
			if err := videoMgr.Stop(); err != nil {
				fmt.Printf("❌ Error stopping video: %v\n", err)
			}
			if err := audioMgr.Stop(); err != nil {
				fmt.Printf("❌ Error stopping audio: %v\n", err)
			} else {
				fmt.Println("✅ Audio module unloaded successfully")
			}
			cleanupDone <- true
		}()

		// Start video
		if err := videoMgr.Start(); err != nil {
			fmt.Printf("❌ %v\n", err)
			audioMgr.Stop()
			os.Exit(1)
		}

		fmt.Println("✨ Running! Press Ctrl+C in this terminal to exit.")

		// Wait for video to finish or signal
		go func() {
			videoMgr.Wait()
			sigChan <- syscall.SIGTERM
		}()

		<-cleanupDone
		fmt.Println("👋 Goodbye!")
	},
}

func init() {
	runCmd.Flags().StringVar(&audioSource, "audio-source", "", "Audio source device (index from list-audio or device name)")
	runCmd.Flags().StringVar(&videoSource, "video-source", "", "Video device (index from list-video or device name, optional, defaults to first available)")
	runCmd.Flags().StringVar(&videoSize, "video-size", "1920x1080", "Video resolution (default: 1920x1080)")
	runCmd.Flags().StringVar(&inputFormat, "input-format", "mjpeg", "Video input format (default: mjpeg)")
	runCmd.Flags().StringVar(&windowTitle, "window-title", "Capture Clam", "Window title (default: Capture Clam)")
	runCmd.Flags().BoolVar(&disableAlwaysOnTop, "no-always-on-top", false, "Disable always-on-top window behavior")
	runCmd.MarkFlagRequired("audio-source")
}
