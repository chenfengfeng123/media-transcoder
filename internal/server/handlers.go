package server

import (
	"context"
	"database/sql"
	"fmt"
	config "github.com/harisbeha/media-transcoder/internal/config"
	"github.com/harisbeha/media-transcoder/internal/data"
	"github.com/harisbeha/media-transcoder/internal/models"
	"github.com/gocraft/work"
	"github.com/labstack/echo/v4"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
	"strconv"
	"sync"
)

type singleResponse struct {
	Message string      `json:"message"`
	Status  int         `json:"status"`
	Job     *models.Job `json:"job"`
}

type multiResponse struct {
	Message string        `json:"message"`
	Status  int           `json:"status"`
	Count   int           `json:"count"`
	Jobs    *[]models.Job `json:"jobs"`
	Page    string        `json:"page"`
}

type request struct {
	Profile     string `json:"profile" binding:"required"`
	Source      string `json:"source" binding:"required"`
	Destination string `json:"dest" binding:"required"`
	C24JobId    string `json:"c24_job_id" binding:"required"`
}

type updateRequest struct {
	Status string `json:"status"`
}

type index struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Docs    string `json:"docs"`
	Github  string `json:"github"`
}

type H map[string]interface{}

func indexHandler(c echo.Context) error {
	resp := index{
		Name:    "c24-media",
		Version: "0.0.1",
		Github:  "https://github.com/harisbeha/media-transcoder",
	}
	return c.JSON(200, resp)
}

func CreateJob(c echo.Context) error {

	req := new(request)
	if err := c.Bind(req); err != nil {
		return err
	}

	// Create Job and push the work to work queue.
	job := models.Job{
		GUID:        xid.New().String(),
		C24JobID:    req.C24JobId,
		Meta: 		 models.JobMetadata{},
		Profile:     req.Profile,
		Source:      req.Source,
		Destination: req.Destination,
		Status:      models.JobQueued, // Status queued.
	}

	// Send to work queue.
	_, err := enqueuer.Enqueue(config.Get().DownloadWorkerJobName, work.Q{
		"guid":        job.GUID,
		"profile":     job.Profile,
		"source":      job.Source,
		"destination": job.Destination,
		"c24_job_id": job.C24JobID,
	})
	if err != nil {
		log.Fatal(err)
	}
	created := data.CreateJob(job)

	// Create the encode relationship.
	ed := models.EncodeData{
		JobID: created.ID,
		Progress: models.NullFloat64{
			NullFloat64: sql.NullFloat64{
				Float64: 0,
				Valid:   true,
			},
		},
		Data: models.NullString{
			NullString: sql.NullString{
				String: "{}",
				Valid:  true,
			},
		},
	}
	edCreated := data.CreateEncodeData(ed)
	created.EncodeDataID = edCreated.EncodeDataID
	log.Info(job)
	return c.JSON(http.StatusOK, H{
		"status": http.StatusOK,
		"message": "OK",
	})
}

func getJobsHandler(c echo.Context) error {
	page := c.QueryParam("page")
	count := c.QueryParam("count")
	pageParam := 0
	countParam := 25
	if page != "" {
		pageParam, _ = strconv.Atoi(page)
	}
	if count != "" {
		countParam, _ = strconv.Atoi(count)
	}

	var wg sync.WaitGroup
	var jobs *[]models.Job
	var jobsCount int

	wg.Add(1)
	go func() {
		jobs = data.GetJobs((pageParam-1)*countParam, countParam)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		jobsCount = data.GetJobsCount()
		wg.Done()
	}()
	wg.Wait()

	return c.JSON(http.StatusOK, H{
		"count": jobsCount,
		"items": jobs,
	})
}

func getJobsByIDHandler(c echo.Context) error {
	id := c.Param("id")
	jobInt, _ := strconv.Atoi(id)

	job, err := data.GetJobByID(jobInt)

	if err != nil {
		log.Print(err.Error())
		return c.JSON(http.StatusNotFound, H{
			"status":  http.StatusNotFound,
			"message": "Job does not exist",
		})
	}

	return c.JSON(http.StatusOK, H{
		"status": http.StatusOK,
		"job":    job,
	})
}

func updateJobByIDHandler(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	// Decode json.
	jsonData := new(updateRequest)
	if err := c.Bind(&jsonData); err != nil {
		errResp := singleResponse{
			Message: err.Error(),
			Status:  500,
			Job:     nil,
		}
		return c.JSON(http.StatusBadRequest, errResp)
	}

	job, err := data.GetJobByID(id)

	resp := singleResponse{
		Message: "Job does not exist",
		Status:  http.StatusNotFound,
		Job:     nil,
	}

	if err != nil {
		return c.JSON(http.StatusNotFound, resp)
	}

	if jsonData.Status != "" {
		job.Status = jsonData.Status
	}

	updatedJob := data.UpdateJobByID(id, *job)
	return c.JSON(http.StatusOK, updatedJob)
}

func getStatsHandler(c echo.Context) error {
	stats, err := data.GetJobsStats()

	if err != nil {
		return c.JSON(http.StatusNotFound, H{
			"status":  http.StatusNotFound,
			"message": "Error getting stats",
		})
	}

	return c.JSON(http.StatusOK, H{
		"status": http.StatusOK,
		"stats": H{
			"jobs": stats,
		},
	})
}

func workerQueueHandler(c echo.Context) error {
	transcodeClient := work.NewClient(config.Get().TranscodeWorkerNamespace, redisPool)
	downloadClient := work.NewClient(config.Get().DownloadWorkerNamespace, redisPool)

	transcodeQueues, err := transcodeClient.Queues()
	downloadQueues, err := downloadClient.Queues()
	respSize := len(transcodeQueues) + len(downloadQueues)
	queues := make([]*work.Queue, 0, respSize)
	log.Info(queues)

	if err != nil {
		fmt.Println(err)
	}
	return c.JSON(200, transcodeQueues)
}

func workerPoolsHandler(c echo.Context) error {
	client := work.NewClient(config.Get().TranscodeWorkerNamespace, redisPool)

	resp, err := client.WorkerPoolHeartbeats()
	if err != nil {
		fmt.Println(err)
	}
	return c.JSON(200, resp)
}

func workerBusyHandler(c echo.Context) error {
	client := work.NewClient(config.Get().TranscodeWorkerNamespace, redisPool)

	observations, err := client.WorkerObservations()
	if err != nil {
		fmt.Println(err)
	}

	var busyObservations []*work.WorkerObservation
	for _, ob := range observations {
		if ob.IsBusy {
			busyObservations = append(busyObservations, ob)
		}
	}
	return c.JSON(200, busyObservations)
}

type s3ListResponse struct {
	Folders []string `json:"folders"`
	Files   []file   `json:"files"`
}

type file struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

func s3ListHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, "")
}

//func s3ListHandler(c echo.Context) error {
//	prefix := c.QueryParam("prefix")
//	files, err := storage.S3ListFiles(prefix)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	resp := s3ListResponse{}
//
//	// var prefixes &[]S3ListResponse.Folders
//	for _, item := range files.CommonPrefixes {
//		resp.Folders = append(resp.Folders, *item.Prefix)
//	}
//
//	for _, item := range files.Contents {
//		var obj file
//		obj.Name = *item.Key
//		obj.Size = *item.Size
//		resp.Files = append(resp.Files, obj)
//	}
//
//	return c.JSON(http.StatusOK, resp)
//}

func profilesHandler(c echo.Context) error {
	profiles := config.Get().Profiles
	return c.JSON(200, H{
		"profiles": profiles,
	})
}

func machinesHandler(c echo.Context) error {
	ctx := context.TODO()

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	fmt.Print(ctx)

	return c.JSON(200, H{
		"machines": nodes,
	})
}

func createMachineHandler(c echo.Context) error {
	// Decode json.
	var json map[string]interface{}
	ctx := context.TODO()
	fmt.Print(ctx, json)

	return c.JSON(200, H{
		"machine": json,
	})
}

func deleteMachineHandler(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var json map[string]interface{}
	ctx := context.TODO()
	fmt.Print(ctx, id, json)

	return c.JSON(200, H{
		"machine": json,
	})
}

func deleteMachineByTagHandler(c echo.Context) error {
	ctx := context.TODO()
	fmt.Print(ctx)

	return c.JSON(200, H{
		"deleted": true,
	})
}

func listMachineRegionsHandler(c echo.Context) error {
	var json map[string]interface{}
	ctx := context.TODO()
	fmt.Print(ctx, json)

	return c.JSON(200, H{
		"regions": json,
	})
}

func listMachineSizesHandler(c echo.Context) error {
	var json map[string]interface{}
	ctx := context.TODO()
	fmt.Print(ctx, json)

	return c.JSON(200, H{
		"sizes": json,
	})
}
