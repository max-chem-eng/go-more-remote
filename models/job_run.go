package models

import (
	"time"

	"gorm.io/gorm"
)

type JobRun struct {
	gorm.Model
	JobID       uint       `json:"job_id" gorm:"column:job_id;not null"`
	StartedAt   *time.Time `json:"started_at" gorm:"column:started_at;"`
	CompletedAt *time.Time `json:"completed_at" gorm:"column:completed_at;"`
	Status      string     `json:"status" gorm:"column:status;not null"`
	Logs        string     `json:"logs" gorm:"column:logs;not null"`
	ExitCode    int        `json:"exit_code" gorm:"column:exit_code"`
	Stats       string     `json:"stats" gorm:"column:stats"`
}

type JobRuns []Job

// var statuses = []string{"running", "completed", "failed"}

func (j *JobRun) Save() error {
	return Db.Save(j).Error
}

func GetJobRuns(jobID string) (JobRuns, error) {
	var runs JobRuns
	err := Db.Where("job_id = ?", jobID).Find(&runs).Error
	return runs, err
}
