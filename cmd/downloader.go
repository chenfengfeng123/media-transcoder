package cmd

import (
	"fmt"
	"github.com/harisbeha/media-transcoder/internal/service"

	config "github.com/harisbeha/media-transcoder/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(downloaderCmd)
}

var downloaderCmd = &cobra.Command{
	Use:   "downloader",
	Short: "Start the download worker.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting downloader...")
		startDownloadWorkers()
	},
}

func startDownloadWorkers() {

	// Worker config.
	workerCfg := &service.WorkerConfig{
		Host:        config.Get().RedisHost,
		Port:        config.Get().RedisPort,
		Namespace:   config.Get().DownloadWorkerNamespace,
		JobName:     config.Get().DownloadWorkerJobName,
		Concurrency: config.Get().WorkerConcurrency,
	}

	// Create Workers.
	service.NewDownloadWorker(*workerCfg)
}
