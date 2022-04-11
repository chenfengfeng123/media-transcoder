package helpers

import (
	"fmt"
	"errors"
	"os"
	"path"
	"math"
	log "github.com/sirupsen/logrus"
)

func CreateLocalSourcePath(workDir string, src string, ID string) string {
	// Get local destination path.
	tmpDir := workDir + "/" + ID + "/"
	os.MkdirAll(tmpDir, 0700)
	os.MkdirAll(tmpDir+"src", 0700)
	os.MkdirAll(tmpDir+"dst", 0700)
	return tmpDir + path.Base(src)
}

func GetTmpPath(workDir string, ID string) string {
	tmpDir := workDir + "/" + ID + "/"
	return tmpDir
}

//ToTimeFormat return string in format time hh:mm:ss
func MsToTimeFormat(duration int) string {
	hours := "00"
	minutes := "00"
	seconds := "00"
	millis := "000"
	ms := 000

	durationMs := duration
	duration = duration / 1000

	if h := int(float64(duration) / 3600); h > 0 {
		if h < 10 {
			hours = fmt.Sprintf("0%d", h)
		} else {
			hours = fmt.Sprintf("%d", h)
		}
	}

	if m := int(float64(duration)/60) % 60; m > 0 {
		if m < 10 {
			minutes = fmt.Sprintf("0%d", m)
		} else {
			minutes = fmt.Sprintf("%d", m)
		}
	}

	if s := duration % 60; s > 0 {
		if s < 10 {
			seconds = fmt.Sprintf("0%d", s)
		} else {
			seconds = fmt.Sprintf("%d", s)
		}
	}

	secs := duration % 60
	secsToMilli := secs * 1000
	ms = int(math.Abs(float64(secsToMilli) - float64(durationMs)))
	millis = fmt.Sprintf("%v", ms)

	return fmt.Sprintf("%s:%s:%s.%s", hours, minutes, seconds, millis)
}

func IsDirectory(path string) bool {
	fd, err := os.Stat(path)
	if err != nil {
		log.Error(err)
		os.Exit(2)
	}
	switch mode := fd.Mode(); {
	case mode.IsDir():
		return true
	case mode.IsRegular():
		return false
	}
	return false
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func FileExists(filename string) error {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return errors.New("File does not exist" + filename)
	}
	if info.IsDir() {
		return errors.New(filename + "is a directory")
	}
	return nil
}