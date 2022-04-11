package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"

	config "github.com/harisbeha/media-transcoder/internal/config"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "c24_media",
	Short: "Onex Labs Media Processing Suite",
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "default", "Config YAML")

	config.LoadConfig(cfgFile)
	fmt.Println(config.Get())
}

// Execute starts cmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}