package storage

import (
	"bytes"
	"os/exec"
	log "github.com/sirupsen/logrus"
)

func gsUtil(args ...string) error {
	cmd := exec.Command("gsutil", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func UploadFile(path string, gsURL string) error {
	if path != "" {
		log.Print(path, gsURL)
		err := gsUtil("cp", path, gsURL)
		if err != nil {
			return err
		}
	}
	return nil
}

func DownloadFile(srcUrl string, path string) error {
	if path != "" {
		log.Print(srcUrl, path)
		err := gsUtil("cp", srcUrl, path)
		if err != nil {
			return err
		}
	}
	return nil
}