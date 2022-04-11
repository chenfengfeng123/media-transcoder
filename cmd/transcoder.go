package cmd

import (
	"fmt"
	"github.com/harisbeha/media-transcoder/internal/service"

	config "github.com/harisbeha/media-transcoder/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(transcoderCmd)
}

var transcoderCmd = &cobra.Command{
	Use:   "transcoder",
	Short: "Start the transcode worker.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting transcoder...")
		startTranscodeWorkers()
	},
}

func startTranscodeWorkers() {

	// Worker config.
	workerCfg := &service.WorkerConfig{
		Host:        config.Get().RedisHost,
		Port:        config.Get().RedisPort,
		Namespace:   config.Get().TranscodeWorkerNamespace,
		JobName:     config.Get().TranscodeWorkerJobName,
		Concurrency: config.Get().WorkerConcurrency,
	}

	// Create Workers.
	service.NewTranscodeWorker(*workerCfg)
}
