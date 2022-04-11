package server

import (
	"fmt"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo/v4"
	"net/http"
)

var (
	redisPool *redis.Pool
	enqueuer  *work.Enqueuer
)

// Config defines configuration for creating a NewServer.
type Config struct {
	ServerPort  string
	RedisHost   string
	RedisPort   int
	Namespace   string
	JobName     string
	Concurrency uint
}

// NewServer creates a new server
func NewServer(serverCfg Config) {
	// Setup redis queue.
	redisPool = &redis.Pool{
		MaxActive: 5,
		MaxIdle:   5,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(fmt.Sprintf("%s:%d", serverCfg.RedisHost, serverCfg.RedisPort))
		},
	}
	enqueuer = work.NewEnqueuer(serverCfg.Namespace, redisPool)

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/dashboard")
	})

	e.File("/dashboard", "/web/dist/index.html")

	// Web dashboard.
	e.Static("/static", "/web/dist")


	//// Catch all fallback for HTML5 History Mode.
	//// https://router.vuejs.org/guide/essentials/history-mode.html
	//e.NoRoute(func(c *gin.Context) {
	//	c.File("./web/dist/index.html")
	//})

	// API routes.
	api := e.Group("/api")
	{
		// Index.
		api.GET("/", indexHandler)

		// S3.
		api.GET("/s3/list", s3ListHandler)

		// Profiles.
		api.GET("/profiles", profilesHandler)

		// Jobs.
		api.POST("/jobs", CreateJob)
		api.GET("/jobs", getJobsHandler)
		api.GET("/jobs/:id", getJobsByIDHandler)
		api.PUT("/jobs/:id", updateJobByIDHandler)

		// Stats.
		api.GET("/stats", getStatsHandler)

		// Worker info.
		api.GET("/worker/queue", workerQueueHandler)
		api.GET("/worker/pools", workerPoolsHandler)
		api.GET("/worker/busy", workerBusyHandler)

		// Machines.
		api.GET("/machines", machinesHandler)
		api.POST("/machines", createMachineHandler)
		api.DELETE("/machines", deleteMachineByTagHandler)
		api.DELETE("/machines/:id", deleteMachineHandler)
		api.GET("/machines/regions", listMachineRegionsHandler)
		api.GET("/machines/sizes", listMachineSizesHandler)
	}
	e.Logger.Fatal(e.Start(":8080"))
}
