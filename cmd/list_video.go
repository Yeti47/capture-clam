package cmd

import (
	"fmt"
	"os"

	"github.com/Yeti47/capture-clam/internal/capture"

	"github.com/spf13/cobra"
)

var listVideoCmd = &cobra.Command{
	Use:   "list-video",
	Short: "List available video devices",
	Long:  `List all available V4L2 video devices.`,
	Run: func(cmd *cobra.Command, args []string) {
		videoDevices, err := capture.ListVideoDevices()
		if err != nil {
			fmt.Printf("❌ Error listing video devices: %v\n", err)
			os.Exit(1)
		}

		if len(videoDevices) == 0 {
			fmt.Println("No video devices found.")
			return
		}

		fmt.Println("Available Video Devices:")
		for i, device := range videoDevices {
			fmt.Printf("%d. %s - %s\n", i+1, device.Device, device.Name)
		}
	},
}
