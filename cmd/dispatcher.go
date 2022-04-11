package cmd

import (
	"fmt"
	"github.com/harisbeha/media-transcoder/internal/subscriber"

	config "github.com/harisbeha/media-transcoder/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(subscriberCmd)
}

var subscriberCmd = &cobra.Command{
	Use:   "dispatcher",
	Short: "Start the dispatcher.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting dispatcher...")
		startSubscribers()
	},
}

func startSubscribers() {

	dispatcherType := config.Get().DispatcherType

	var jobNamespace, jobName string

	if dispatcherType == "download" {
		jobNamespace = config.Get().DownloadWorkerNamespace
		jobName = config.Get().DownloadWorkerJobName

	} else if dispatcherType == "transcode" {
		jobNamespace = config.Get().TranscodeWorkerNamespace
		jobName = config.Get().TranscodeWorkerJobName
	}

	// Subscriber config.
	dispatcherCfg := &subscriber.Config{
		RedisHost:   config.Get().RedisHost,
		RedisPort:   config.Get().RedisPort,
		Namespace:   jobNamespace,
		JobName:     jobName,
		Concurrency: config.Get().WorkerConcurrency,
	}

	// Create Workers.
	subscriber.NewSubscriber(*dispatcherCfg)
}
