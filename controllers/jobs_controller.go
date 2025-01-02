package controllers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/max-chem-eng/go-more-remote/jobengine"
	"github.com/max-chem-eng/go-more-remote/models"
)

type JobsController struct {
	BaseController
}

func NewJobsController() Controller {
	return &JobsController{}
}

func (jc *JobsController) SetupRoutes(r *gin.Engine) {
	JobsRoutes := r.Group("/jobs")
	JobsRoutes.GET("", jc.Index)
	JobsRoutes.GET("/:id", jc.ShowJob)
	JobsRoutes.POST("", jc.CreateJob)
	JobsRoutes.DELETE("/:id", jc.DeleteJob)
	JobsRoutes.POST("/:id/execute", jc.ExecuteJob)
}

func (jc *JobsController) Index(c *gin.Context) {
	jobs, err := models.GetAllJobs()
	if err != nil {
		handleError(c, StatusInternalServerError, "Error getting jobs", err)
		return
	}

	c.IndentedJSON(StatusOK, jobs)
}

func (jc *JobsController) ShowJob(c *gin.Context) {
	jobIDInt, _ := strconv.Atoi(c.Param("id"))
	jobID := uint(jobIDInt)

	job, err := models.GetJob(uint(jobID))
	if err != nil {
		handleError(c, StatusInternalServerError, "Error getting job", err)
		return
	}

	c.IndentedJSON(StatusOK, job)
}

func (jc *JobsController) CreateJob(c *gin.Context) {
	var job models.Job
	if err := c.ShouldBindJSON(&job); err != nil {
		handleError(c, StatusInternalServerError, "Error binding job", err)
		return
	}

	if err := job.Save(); err != nil {
		handleError(c, StatusInternalServerError, "Error saving job", err)
		return
	}

	c.IndentedJSON(StatusOK, job)
}

func (jc *JobsController) DeleteJob(c *gin.Context) {
	jobIDInt, _ := strconv.Atoi(c.Param("id"))
	jobID := uint(jobIDInt)

	job, err := models.GetJob(uint(jobID))
	if err != nil {
		handleError(c, StatusInternalServerError, "Error getting job", err)
		return
	}

	if err := job.Delete(); err != nil {
		handleError(c, StatusInternalServerError, "Error deleting job", err)
		return
	}

	c.IndentedJSON(StatusOK, gin.H{
		"message": fmt.Sprintf("Job %d deleted", jobID),
	})
}

func (jc *JobsController) ExecuteJob(c *gin.Context) {
	jobIDInt, _ := strconv.Atoi(c.Param("id"))
	jobID := uint(jobIDInt)

	job, err := models.GetJob(uint(jobID))
	if err != nil {
		handleError(c, StatusInternalServerError, "Error getting job", err)
		return
	}

	run := models.JobRun{
		JobID:  jobID,
		Status: "queued",
	}
	if err := run.Save(); err != nil {
		handleError(c, StatusInternalServerError, "Error creating job run", err)
		return
	}

	run.Status = "running"
	now := time.Now()
	run.StartedAt = &now
	if err := run.Save(); err != nil {
		handleError(c, StatusInternalServerError, "Error updating job run status", err)
		return
	}

	go func(j *models.Job, r *models.JobRun) {
		logs, stats, err := jobengine.ExecuteJob(jobengine.JobConfig{
			Language:      j.Language,
			ScriptContent: job.ScriptContent,
			Image:         job.Image,
		})

		now = time.Now()
		run.CompletedAt = &now

		if err != nil {
			run.Status = "failed"
			run.Logs = fmt.Sprintf("Error: %v", err)
		} else {
			run.Status = "completed"
			run.Logs = logs
			run.Stats = stats
		}
		if err := run.Save(); err != nil {
			handleError(c, StatusInternalServerError, "Error saving job run logs", err)
			return
		}
	}(job, &run)

	c.JSON(StatusOK, run)
}
