package controllers

import (
	"github.com/gin-gonic/gin"
)

type HealthCheckController struct {
	BaseController
}

func NewHealthCheckController() Controller {
	return &HealthCheckController{}
}

func (hc *HealthCheckController) SetupRoutes(r *gin.Engine) {
	healthCheckRoutes := r.Group("/up")
	healthCheckRoutes.GET("", hc.HealthCheck)
}

func (hc *HealthCheckController) HealthCheck(c *gin.Context) {
	c.Status(StatusOK)
}
