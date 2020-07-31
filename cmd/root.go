package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
	initHueCmd()
}

var rootCmd = &cobra.Command{
	Use:   "erikbotdev",
	Short: "Twitch Bot",
	Long:  `Twitch bot for ErikDotDev`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
