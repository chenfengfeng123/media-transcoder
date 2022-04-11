package data

import (
	"fmt"
	models "github.com/harisbeha/media-transcoder/internal/models"
)

// GetJobs Gets all jobs.
func GetJobs(offset, count int) *[]models.Job {
	const query = `
	  SELECT
        jobs.*,
        transcode.id "transcode.id",
        transcode.data "transcode.data",
        transcode.progress "transcode.progress"
	  FROM jobs
      LEFT JOIN transcode ON jobs.id = transcode.job_id
	  ORDER BY id DESC
      LIMIT $1 OFFSET $2`

	db, _ := ConnectDB()
	jobs := []models.Job{}
	err := db.Select(&jobs, query, count, offset)
	if err != nil {
		fmt.Println(err)
	}
	db.Close()
	return &jobs
}

// GetJobByID Gets a job by ID.
func GetJobByID(id int) (*models.Job, error) {
	const query = `
      SELECT
        jobs.*,
        transcode.id "transcode.id",
        transcode.data "transcode.data",
        transcode.progress "transcode.progress"
      FROM jobs
      LEFT JOIN transcode ON jobs.id = transcode.job_id
      WHERE jobs.id = $1`

	db, _ := ConnectDB()
	job := models.Job{}
	err := db.Get(&job, query, id)
	if err != nil {
		fmt.Println(err)
		return &job, err
	}
	db.Close()
	return &job, nil
}

// GetJobByGUID Gets a job by GUID.
func GetJobByGUID(id string) (*models.Job, error) {
	const query = `
      SELECT
        jobs.*,
        transcode.id "transcode.id",
        transcode.data "transcode.data",
        transcode.progress "transcode.progress"
      FROM jobs
      LEFT JOIN transcode ON jobs.id = transcode.job_id
      WHERE jobs.guid = $1`

	db, _ := ConnectDB()
	job := models.Job{}
	err := db.Get(&job, query, id)
	if err != nil {
		fmt.Println(err)
		return &job, err
	}
	db.Close()
	return &job, nil
}

// GetJobsCount Gets a count of all jobs.
func GetJobsCount() int {
	var count int
	const query = `SELECT COUNT(*) FROM jobs`

	db, _ := ConnectDB()
	err := db.Get(&count, query)
	if err != nil {
		fmt.Println(err)
	}
	db.Close()
	return count
}

// Stats struct for displaying status and count of a job.
type Stats struct {
	Status string `db:"status" json:"status"`
	Count  int    `db:"count" json:"count"`
}

// GetJobsStats Gets a count of each status.
func GetJobsStats() (*[]Stats, error) {
	const query = `SELECT status, count(status) FROM jobs GROUP BY status, status;`

	s := []Stats{}
	db, _ := ConnectDB()
	err := db.Select(&s, query)
	if err != nil {
		fmt.Println(err)
		return &s, err
	}
	db.Close()

	// Set all statuses.
	var resp []Stats
	for _, v := range models.JobStatuses {
		r := Stats{}
		for _, j := range s {
			if j.Status == v {
				r.Status = j.Status
				r.Count = j.Count
			} else {
				r.Status = v
			}
		}
		resp = append(resp, r)
	}
	return &resp, nil
}

// CreateJob creates a job in database.
func CreateJob(job models.Job) *models.Job {
	const query = `
      INSERT INTO
        jobs (guid,profile,status,c24_job_id,action,source,destination)
      VALUES (:guid,:profile,:status,:c24_job_id,:action,:source,:destination)
      RETURNING id`

	db, _ := ConnectDB()
	tx := db.MustBegin()
	stmt, err := tx.PrepareNamed(query)
	if err != nil {
		fmt.Println("Error", err.Error())
	}

	var id int64 // Returned ID.
	err = stmt.QueryRowx(&job).Scan(&id)
	if err != nil {
		fmt.Println("Error", err.Error())
	}
	tx.Commit()

	// Set to Job type response.
	job.ID = id

	db.Close()
	return &job
}

// CreateEncodeData creates transcode in database.
func CreateEncodeData(ed models.EncodeData) *models.EncodeData {
	const query = `
      INSERT INTO
        transcode (data,progress,job_id)
      VALUES (:data,:progress,:job_id)
      RETURNING id`

	db, _ := ConnectDB()
	tx := db.MustBegin()
	stmt, err := tx.PrepareNamed(query)
	if err != nil {
		fmt.Println("Error", err.Error())
	}

	var id int64 // Returned ID.
	err = stmt.QueryRowx(&ed).Scan(&id)
	if err != nil {
		fmt.Println("Error", err.Error())
	}
	tx.Commit()

	// Set to Job type response.
	ed.EncodeDataID = id

	db.Close()
	return &ed
}

// UpdateEncodeDataByID Update transcode by ID.
func UpdateEncodeDataByID(id int64, jsonString string) error {
	const query = `UPDATE transcode SET data = $1 WHERE id = $2`

	db, _ := ConnectDB()
	tx := db.MustBegin()
	_, err := tx.Exec(query, jsonString, id)
	if err != nil {
		fmt.Println(err)
		return err
	}
	tx.Commit()

	db.Close()
	return nil
}

// UpdateEncodeProgressByID Update progress by ID.
func UpdateEncodeProgressByID(id int64, progress float64) error {
	const query = `UPDATE transcode SET progress = $1 WHERE id = $2`

	db, _ := ConnectDB()
	tx := db.MustBegin()
	_, err := tx.Exec(query, progress, id)
	if err != nil {
		fmt.Println(err)
		return err
	}
	tx.Commit()

	db.Close()
	return nil
}

// UpdateJobByID Update job by ID.
func UpdateJobByID(id int, job models.Job) *models.Job {
	const query = `UPDATE jobs SET status = :status WHERE id = :id`

	db, _ := ConnectDB()
	tx := db.MustBegin()
	_, err := tx.NamedExec(query, &job)
	if err != nil {
		fmt.Println(err)
	}
	tx.Commit()

	db.Close()
	return &job
}

// UpdateJobStatus Update job status by ID.
func UpdateJobStatus(guid string, status string) error {
	const query = `UPDATE jobs SET status = $1 WHERE guid = $2`

	db, _ := ConnectDB()
	tx := db.MustBegin()
	_, err := tx.Exec(query, status, guid)
	if err != nil {
		fmt.Println(err)
		return err
	}
	tx.Commit()

	db.Close()
	return nil
}