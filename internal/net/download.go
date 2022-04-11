package net

import (
	"github.com/harisbeha/media-transcoder/internal/models"
	storage "github.com/harisbeha/media-transcoder/internal/storage"
)

// DownloadFunc creates a download.
type DownloadFunc func(job models.Job) error

// GetDownloader sets the download function.
func GetDownloader() *storage.S3 {
	return &storage.S3{}
}