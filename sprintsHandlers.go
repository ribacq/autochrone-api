package main

import (
	"github.com/gin-gonic/gin"

	"fmt"
	"net/http"
	"time"
)

// SprintsGET responds with a project’s sprints
func SprintsGET(c *gin.Context) {
	project := c.MustGet("project").(*Project)

	if err := project.FetchSprints(); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, project.Sprints)
}

// SprintRequest determines fields for a sprint request
type SprintRequest struct {
	TimeStart string `json:"timeStart"`
	Duration  int    `json:"duration"`
	Break     int    `json:"break"`
}

// SprintsPOST saves a given sprint and returns its API location
// requires json(timeStart, duration, break)
func SprintsPOST(c *gin.Context) {
	project := c.MustGet("project").(*Project)

	req := &SprintRequest{}
	if err := c.BindJSON(req); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	timeStart, err := time.Parse("2006-01-02 15:04:05", req.TimeStart)
	if err != nil || req.Duration < 1 || req.Duration < 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	sprint, err := project.NewSprint(timeStart, req.Duration, req.Break)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header("Location", fmt.Sprintf("/sprints/%s", sprint.Slug))
	c.Status(http.StatusOK)
}
