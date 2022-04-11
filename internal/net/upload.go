package net

import (
	"github.com/harisbeha/media-transcoder/internal/storage"
	"github.com/harisbeha/media-transcoder/internal/models"

)

// UploadFunc creates a upload.
type UploadFunc func(job models.Job) error

// GetUploader gets the upload function.
func GetUploader() *storage.S3 {
	return &storage.S3{}
}