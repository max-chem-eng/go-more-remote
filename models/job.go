package models

import (
	"gorm.io/gorm"
)

type Job struct {
	gorm.Model
	Language      string   `json:"language" gorm:"column:language;not null"`
	Image         string   `json:"image" gorm:"column:image"`
	ScriptContent string   `json:"script_content" gorm:"column:script_content;not null"`
	JobRuns       []JobRun `json:"job_runs" gorm:"foreignKey:JobID"`
}

type Jobs []Job

func (j *Job) Save() error {
	return Db.Save(j).Error
}

func (j *Job) Delete() error {
	err := Db.Where("job_id = ?", j.ID).Delete(&JobRun{}).Error
	if err != nil {
		return err
	}
	return Db.Delete(j).Error
}

// func DeleteJob(id uint) error {
// 	err := Db.Where("job_id = ?", j.ID).Delete(&JobRun{}).Error
// 	if err != nil {
// 		return err
// 	}
// 	return Db.Where("id = ?", id).Delete(&Job{}).Error
// }

func GetJob(id uint) (*Job, error) {
	var job Job
	err := Db.Preload("JobRuns").First(&job, id).Error
	return &job, err
}

func GetAllJobs() (Jobs, error) {
	var jobs Jobs
	err := Db.Find(&jobs).Error
	return jobs, err
}
