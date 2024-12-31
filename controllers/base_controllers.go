package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	StatusOK                  = http.StatusOK
	StatusInternalServerError = http.StatusInternalServerError
	StatusBadRequest          = http.StatusBadRequest
)

type BaseController struct {
}

type Controller interface {
	SetupRoutes(r *gin.Engine)
}

func SetupRoutes(r *gin.Engine) {
	NewHealthCheckController().SetupRoutes(r)
	NewJobsController().SetupRoutes(r)
}

func handleError(c *gin.Context, status int, message string, err error) {
	c.IndentedJSON(status, gin.H{
		"message": message,
		"error":   err.Error(),
	})
}
