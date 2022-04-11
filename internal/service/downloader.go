package service

import (
	"fmt"
	"github.com/harisbeha/media-transcoder/internal/actions"
	_ "github.com/harisbeha/media-transcoder/internal/actions"
	config "github.com/harisbeha/media-transcoder/internal/config"
	"github.com/harisbeha/media-transcoder/internal/models"
	_ "github.com/harisbeha/media-transcoder/internal/models"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
)

// NewWorker creates a new worker instance to listen and process jobs in the queue.
func NewDownloadWorker(workerCfg WorkerConfig) {

	// Make a redis pool
	redisPool := &redis.Pool{
		MaxActive: 5,
		MaxIdle:   5,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(fmt.Sprintf("%s:%d", workerCfg.Host, workerCfg.Port))
		},
	}

	// Make a new pool.
	pool := work.NewWorkerPool(Context{},
		workerCfg.Concurrency, workerCfg.Namespace, redisPool)

	// Add middleware that will be executed for each job
	pool.Middleware((*Context).Log)
	pool.Middleware((*Context).FindJob)

	// Map the name of jobs to handler functions
	pool.Job(config.Get().DownloadWorkerJobName, (*Context).SendDownloadJob)

	// Customize options:
	// pool.JobWithOptions("export", work.JobOptions{Priority: 10, MaxFails: 1}, (*Context).Export)

	// Start processing jobs
	pool.Start()

	// Wait for a signal to quit:
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan

	// Stop the pool
	pool.Stop()
}

// SendJob worker handler for running job.
func (c *Context) SendDownloadJob(job *work.Job) error {
	guid := job.ArgString("guid")
	profile := job.ArgString("profile")
	source := job.ArgString("source")
	destination := job.ArgString("destination")
	c24JobID := job.ArgString("c24_job_id")

	j := models.Job{
		GUID:        guid,
		C24JobID: 	 c24JobID,
		Profile:     profile,
		Source:      source,
		Destination: destination,
	}

	// Start job.
	actions.RunDownloadJob(j)
	log.Infof("worker: completed %s!\n", j.Profile)
	//defer os.Exit(0)
	return nil
}