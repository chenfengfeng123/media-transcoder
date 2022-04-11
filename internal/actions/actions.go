package actions

import (
	"encoding/json"
	"fmt"
	"github.com/harisbeha/media-transcoder/internal/alert"
	config "github.com/harisbeha/media-transcoder/internal/config"
	data "github.com/harisbeha/media-transcoder/internal/data"
	"github.com/harisbeha/media-transcoder/internal/helpers"
	"github.com/harisbeha/media-transcoder/internal/kube"
	models "github.com/harisbeha/media-transcoder/internal/models"
	ffprobe "github.com/harisbeha/media-transcoder/internal/probe"
	"github.com/harisbeha/media-transcoder/internal/storage"
	transcode "github.com/harisbeha/media-transcoder/internal/transcode"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var progressCh chan struct{}

const progressInterval = time.Second * 2

func download(job models.Job) error {
	log.Info("running download task")

	// Update status.
	data.UpdateJobStatus(job.GUID, models.JobDownloading)

	// Get job data.
	j, _ := data.GetJobByGUID(job.GUID)
	encodeID := j.EncodeDataID

	//// Get downloader type.
	//d := net.GetDownloader()
	//
	//// Do download and track progress.
	//go trackTransferProgress(encodeID, d)
	//err := d.S3Download(job)
	//if err != nil {
	//	log.Error(err)
	//}
	//
	//// Close channel to stop progress updates.
	//close(progressCh)
	storagePath := job.Destination
	if job.Destination == "local" {
		storagePath = fmt.Sprintf("%s/src/%s", config.Get().WorkDirectory, job.C24JobID)
	}
	err := storage.DownloadFile(job.Source, storagePath)
	err = helpers.FileExists(storagePath)
	if err != nil {
		log.Info(err.Error())
	}

	// Set progress to 100.
	data.UpdateEncodeProgressByID(encodeID, 100)
	return err
}

func probe(job models.Job) (*ffprobe.FFProbeResponse, error) {
	log.Info("running probe task")

	// Update status.
	data.UpdateJobStatus(job.GUID, models.JobProbing)

	// Run FFProbe.
	f := ffprobe.FFProbe{}

	sourceMediaPath := getSourceMediaPath(job.C24JobID)
	probeData := f.Run(sourceMediaPath)

	// Add probe data to DB.
	b, err := json.Marshal(probeData)
	if err != nil {
		log.Error(err)
	}
	j, _ := data.GetJobByGUID(job.GUID)
	data.UpdateEncodeDataByID(j.EncodeDataID, string(b))

	return probeData, nil
}

func encode(job models.Job, probeData *ffprobe.FFProbeResponse) error {
	log.Info("running encode task")

	// Update status.
	data.UpdateJobStatus(job.GUID, models.JobEncoding)

	p, err := config.GetFFmpegProfile(job.Profile)
	if err != nil {
		return err
	}
	//dest := path.Dir(job.LocalSource) + "/dst/" + p.Output

	// Get job data.
	j, _ := data.GetJobByGUID(job.GUID)
	encodeID := j.EncodeDataID

	// Run FFmpeg.
	f := &transcode.FFmpeg{}
	go trackEncodeProgress(encodeID, probeData, f)
	sourceMediaPath := getSourceMediaPath(j.C24JobID)
	log.Info("source media path", sourceMediaPath)
	dest := "/mpc/dst/" + j.C24JobID + p.Output
	f.Run(sourceMediaPath, dest, p.Options)
	close(progressCh)

	// Set encode progress to 100.
	data.UpdateEncodeProgressByID(encodeID, 100)
	return err
}

func upload(job models.Job) error {
	log.Info("running upload task")

	// Update status.
	data.UpdateJobStatus(job.GUID, models.JobUploading)

	// Get job data.
	j, _ := data.GetJobByGUID(job.GUID)
	encodeID := j.EncodeDataID

	//d := net.GetUploader()
	//
	//// Do download and track progress.
	//go trackTransferProgress(encodeID, d)
	//err := d.S3Upload(job)
	//if err != nil {
	//	log.Error(err)
	//}
	//
	//// Close channel to stop progress updates.
	//close(progressCh)
	p, err := config.GetFFmpegProfile(job.Profile)
	if err != nil {
		return err
	}
	localPath := "/mpc/dst/" + j.C24JobID + p.Output
	storagePath := localPath
	storageUrl := prefixPathWithService("gs", "dev-experiments", storagePath)
	log.Info(storageUrl)
	err = helpers.FileExists(localPath)
	err = storage.UploadFile(localPath, j.Destination)

	// Set progress to 100.
	data.UpdateEncodeProgressByID(encodeID, 100)
	return err
}

func cleanup(job models.Job) error {
	log.Info("running cleanup task")

	tmpPath := helpers.GetTmpPath(config.Get().WorkDirectory, job.GUID)
	err := os.RemoveAll(tmpPath)
	if err != nil {
		return err
	}
	return nil
}

func completed(job models.Job) error {
	log.Info("job completed")

	// Update status.
	data.UpdateJobStatus(job.GUID, models.JobCompleted)
	return nil
}

func completeDownload(job models.Job) error {
	log.Info("Download completed")
	// Update status.
	data.UpdateJobStatus(job.GUID, models.JobCompleted)
	return nil
}

func sendAlert(job models.Job) error {
	log.Info("sending alert")

	message := fmt.Sprintf(
		"*Encode Successful!* :tada:\n"+
			"*ID*: %s:\n"+
			"*Job ID*: %s:\n"+
			"*Profile*: %s\n"+
			"*Source*: %s\n"+
			"*Destination*: %s\n\n",
		job.GUID, job.C24JobID, job.Profile, job.Source, job.Destination)
	err := alert.SendSlackMessage(config.Get().SlackWebhook, message)
	if err != nil {
		return err
	}
	return nil
}

func notifyCompletion(job models.Job) error {
	log.Info("sending alert")

	message := fmt.Sprintf(
		"*Encode Successful!* :tada:\n"+
			"*Job ID*: %s:\n"+
			"*Profile*: %s\n"+
			"*Source*: %s\n"+
			"*Destination*: %s\n\n",
		job.GUID, job.Profile, job.Source, job.Destination)
	err := alert.SendSlackMessage(config.Get().SlackWebhook, message)
	if err != nil {
		return err
	}
	return nil
}

func RunDownloadJob(job models.Job) {
	//job.LocalSource = helpers.CreateLocalSourcePath(
	//	config.Get().WorkDirectory, localPath, job.GUID)

	// 1. Download.
	err := download(job)
	if err != nil {
		log.Error(err)
		data.UpdateJobStatus(job.GUID, models.JobError)
		return
	}
	completeDownload(job)
	if err != nil {
		log.Error(err)
	}
}

func PrepareEncodeJob(job models.Job) {
	minCpu, maxCpu := getCPULevels()
	minMem, maxMem := getMemoryLevels()
	parallelism := "1"
	maxRetries := "1"
	kubeJob := createKubeJob(job.C24JobID, job.Action, parallelism, maxRetries, minMem, maxMem, minCpu, maxCpu)
	_, err := shipKubeJob(kubeJob)
	if err != nil {
		log.Fatal(err)
	}
}

func RunEncodeJob(job models.Job) {

	sourceMediaPath := getSourceMediaPath(job.C24JobID)
	err := helpers.FileExists(sourceMediaPath)

	// 2. Probe data.
	probeData, err := probe(job)
	if err != nil {
		log.Error(err)
		data.UpdateJobStatus(job.GUID, models.JobError)
		return
	}

	// 3. Encode.
	err = encode(job, probeData)
	if err != nil {
		log.Error(err)
		data.UpdateJobStatus(job.GUID, models.JobError)
		return
	}

	// 4. Upload.
	err = upload(job)
	if err != nil {
		log.Error(err)
		data.UpdateJobStatus(job.GUID, models.JobError)
		return
	}

	// 5. Cleanup.
	//err = cleanup(job)
	//if err != nil {
	//	log.Error(err)
	//	data.UpdateJobStatus(job.GUID, models.JobError)
	//	return
	//}

	// 6. Done
	completed(job)
	if err != nil {
		log.Error(err)
	}

	// 7. Alert
	notifyCompletion(job)
	if err != nil {
		log.Error(err)
	}

	// 7. Alert
	sendAlert(job)
	if err != nil {
		log.Error(err)
	}
}

func trackEncodeProgress(encodeID int64, p *ffprobe.FFProbeResponse, f *transcode.FFmpeg) {
	progressCh = make(chan struct{})
	ticker := time.NewTicker(progressInterval)

	for {
		select {
		case <-progressCh:
			ticker.Stop()
			return
		case <-ticker.C:
			currentFrame := f.Progress.Frame
			totalFrames, _ := strconv.Atoi(p.Streams[0].NbFrames)

			pct := (float64(currentFrame) / float64(totalFrames)) * 100

			// Update DB with progress.
			pct = math.Round(pct*100) / 100
			fmt.Printf("progress: %d / %d - %0.2f%%\r", currentFrame, totalFrames, pct)
			data.UpdateEncodeProgressByID(encodeID, pct)
		}
	}
}

func trackTransferProgress(encodeID int64, d *storage.S3) {
	progressCh = make(chan struct{})
	ticker := time.NewTicker(progressInterval)

	for {
		select {
		case <-progressCh:
			ticker.Stop()
			return
		case <-ticker.C:
			fmt.Println("transfer progress: ", d.Progress.Progress)
			data.UpdateEncodeProgressByID(encodeID, float64(d.Progress.Progress))
		}
	}
}

func createKubeJob(jobName, action, parallelism, maxRetries, memMin, memMax, cpuMin, cpuMax string) *batchv1.Job {

	actionArg := "transcoder"
	if action == "download" {
		actionArg = "downloader"
	}

	j := kube.New(
		kube.WithJobNameEnv(jobName),
		kube.WithCommandArgs(actionArg),
		kube.WithJobName(jobName),
		kube.WithMaxParallelism(parallelism),
		kube.WithMaxRetries(maxRetries),
		kube.WithTTLCleanupTime(int32(0)),
		kube.WithResourceLimit(cpuMax, memMax),
		kube.WithResourceRequest(cpuMin, memMin),
	)
	return j

}

func shipKubeJob(job *batchv1.Job) (batchv1.Job, error) {
	conf, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}
	// create the clientset
	clientset, err := kubernetes.NewForConfig(conf)
	if err != nil {
		log.Fatal(err)
	}
	jobResp, err := clientset.BatchV1().Jobs("jobs").Create(job)
	if err != nil {
		log.Fatal(err)
	}
	log.Info(jobResp.Name, jobResp.Status)
	return *jobResp, nil
}

func getMemoryLevels() (string, string) {
	// TODO: Pull defaults from config
	return "2000M", "2000M"
}

func getCPULevels() (string, string) {
	// TODO: Pull defaults from config
	return "750m", "750m"
}

func getSourceMediaPath(c24JobID string) string {
	log.Info("smedia", c24JobID)
	path := fmt.Sprintf("%s/src/%s", config.Get().WorkDirectory, c24JobID)
	return path
}

func stripServiceFromURL(url string) string {
	formattedUrl := strings.Replace(url, "gs://", "", -1)
	formattedUrl = strings.Replace(formattedUrl, "s3://", "", -1)
	return formattedUrl
}

func prefixPathWithService(service, bucket, path string) string {
	formattedUrl := fmt.Sprintf("%s://%s/%s", service, bucket, path)
	return formattedUrl
}

// Add profile/variant converter
// Create src/ media folder that transcode jobs work from
// Output files from transcode jobs will go to /dst/PATH
// Make sure service prefixes (s3://, gs://) get stripped before
// dealing with local files
// Make sure service prefixes aren't stripped when uploading to a service
