package service

import (
	"fmt"
	"github.com/harisbeha/media-transcoder/internal/models"
	"github.com/gocraft/work"
	log "github.com/sirupsen/logrus"
	"time"
)

// Config defines configuration for creating a NewWorker.
type WorkerConfig struct {
	Host        string
	Port        int
	Namespace   string
	JobName     string
	Concurrency uint
}

// Context defines the job context to be passed to the worker.
type Context struct {
	GUID        string
	Profile     string
	Source      string
	Destination string
}

// Log worker middleware for logging job.
func (c *Context) Log(job *work.Job, next work.NextMiddlewareFunc) error {
	log.Infof("worker: starting job %s\n", job.Name)
	return next()
}

// FindJob worker middleware for setting job context from job arguments.
func (c *Context) FindJob(job *work.Job, next work.NextMiddlewareFunc) error {
	if _, ok := job.Args["guid"]; ok {
		c.GUID = job.ArgString("guid")
		if err := job.ArgError(); err != nil {
			return err
		}
	}
	if _, ok := job.Args["profile"]; ok {
		c.Profile = job.ArgString("profile")
		if err := job.ArgError(); err != nil {
			return err
		}
	}
	if _, ok := job.Args["source"]; ok {
		c.Source = job.ArgString("source")
		if err := job.ArgError(); err != nil {
			return err
		}
	}
	if _, ok := job.Args["destination"]; ok {
		c.Destination = job.ArgString("destination")
		if err := job.ArgError(); err != nil {
			return err
		}
	}
	return next()
}

func startJob(id int, j models.Job) {
	log.Infof("worker: started %s\n", j.Profile)

	// runWorkflow(j)
	log.Infof("worker: completed %s!\n", j.Profile)
}

func (c *Context) Export(job *work.Job) error {
	job.Checkin("i=" + fmt.Sprint(time.Now().String()))   // Here's the magic! This tells gocraft/work our status
	return nil
}
