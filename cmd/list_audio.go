package cmd

import (
	"fmt"
	"os"

	"github.com/Yeti47/capture-clam/internal/devices"

	"github.com/spf13/cobra"
)

var listAudioCmd = &cobra.Command{
	Use:   "list-audio",
	Short: "List available audio sources",
	Long:  `List all available PulseAudio sources for audio input.`,
	Run: func(cmd *cobra.Command, args []string) {
		deviceLister := devices.NewDeviceLister()
		audioDevices, err := deviceLister.ListAudioDevices()
		if err != nil {
			fmt.Printf("❌ Error listing audio devices: %v\n", err)
			os.Exit(1)
		}

		if len(audioDevices) == 0 {
			fmt.Println("No audio devices found.")
			return
		}

		fmt.Println("Available Audio Sources:")
		for i, device := range audioDevices {
			fmt.Printf("%d. %s - %s\n", i+1, device.Name, device.Description)
		}
	},
}
