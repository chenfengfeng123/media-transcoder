package cmd

import (
	"fmt"
	_ "github.com/gocraft/work"
	"runtime"

	config "github.com/harisbeha/media-transcoder/internal/config"
	server "github.com/harisbeha/media-transcoder/internal/server"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the server.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting server...")
		startServer()
	},
}

func configRuntime() {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)
	fmt.Printf("Running with %d CPUs\n", numCPU)
}
func startServer() {

	// Server config.
	serverCfg := &server.Config{
		ServerPort:  config.Get().Port,
		RedisHost:   config.Get().RedisHost,
		RedisPort:   config.Get().RedisPort,
		Namespace:   config.Get().TranscodeWorkerNamespace,
		JobName:     config.Get().TranscodeWorkerJobName,
		Concurrency: config.Get().WorkerConcurrency,
	}

	// Create HTTP Server.
	configRuntime()
	server.NewServer(*serverCfg)
}