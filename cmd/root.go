package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "capture-clam",
	Short: "Ultra-Low Latency Video Capture Application",
	Long:  `A modular video capturing application with audio loopback support.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(listAudioCmd)
	rootCmd.AddCommand(listVideoCmd)
	rootCmd.AddCommand(cleanupCmd)
}
