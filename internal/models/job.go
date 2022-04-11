package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

// Job status types.
const (
	JobQueued      = "queued"
	JobDownloading = "downloading"
	JobDownloaded  = "downloaded"
	JobProbing     = "probing"
	JobEncoding    = "encoding"
	JobUploading   = "uploading"
	JobCompleted   = "completed"
	JobError       = "error"
)

// JobStatuses All job status types.
var JobStatuses = []string{
	JobQueued,
	JobDownloading,
	JobDownloaded,
	JobProbing,
	JobEncoding,
	JobUploading,
	JobCompleted,
	JobError,
}

// Job describes the job info.
type Job struct {
	ID          int64  `db:"id" json:"id"`
	GUID        string `db:"guid" json:"guid"`
	Profile     string `db:"profile" json:"profile"`
	CreatedDate string `db:"created_date" json:"created_date"`
	C24JobID 	string `db:"c24_job_id" json:"c24_job_id"`
	Status      string `db:"status" json:"status"`
	Meta 		JobMetadata `db:"metadata" json:"metadata"`
	Callback 	Callback `db:"callback" json:"callback"`
	Action		string `db:"action" json:"action"`

	// EncodeData.
	EncodeData `db:"transcode"`

	Source           string `db:"source" json:"source,omitempty"`
	Destination      string `db:"destination" json:"destination,omitempty"`
	LocalSource      string `json:"local_source,omitempty"`
	LocalDestination string `json:"local_destination,omitempty"`
}

type JobMetadata map[string]interface{}
type Callback map[string]interface{}

//type Callback struct {
//	Method       string `json:"method"`
//	Action       string `json:"action"`
//	Topic        string `json:"topic"`
//	Subscription string `json:"subscription"`
//}

// EncodeData describes the encode data.
type EncodeData struct {
	EncodeDataID int64       `db:"id" json:"-"`
	JobID        int64       `db:"job_id" json:"-"`
	Data         NullString  `db:"data" json:"encode,omitempty"`
	Progress     NullFloat64 `db:"progress" json:"progress,omitempty"`
}

// NullString is an alias for sql.NullString data type
type NullString struct {
	sql.NullString
}

// MarshalJSON for NullString
func (ns *NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

// NullInt64 is an alias for sql.NullInt64 data type
type NullInt64 struct {
	sql.NullInt64
}

// MarshalJSON for NullInt64
func (ni *NullInt64) MarshalJSON() ([]byte, error) {
	if !ni.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ni.Int64)
}

// NullFloat64 is an alias for sql.NullFloat64 data type
type NullFloat64 struct {
	sql.NullFloat64
}

// MarshalJSON for NullFloat64
func (nf *NullFloat64) MarshalJSON() ([]byte, error) {
	if !nf.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nf.Float64)
}

func (m JobMetadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *JobMetadata) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(b, &m)
}

func (c Callback) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *Callback) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(b, &c)
}