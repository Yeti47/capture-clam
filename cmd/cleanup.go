package cmd

import (
	"fmt"

	"github.com/Yeti47/capture-clam/internal/capture"

	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up rogue audio loopback modules",
	Long:  `Unload all PulseAudio loopback modules that may be left over from interrupted runs.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🧹 Cleaning up audio loopback modules...")

		if err := capture.CleanupLoopbackModules(); err != nil {
			fmt.Printf("❌ Error during cleanup: %v\n", err)
			return
		}

		fmt.Println("✅ All loopback modules unloaded successfully")
		fmt.Println("🧹 Cleanup complete!")
	},
}
