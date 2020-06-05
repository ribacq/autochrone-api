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

// SprintPOSTRequest determines fields for a sprint request
type SprintPOSTRequest struct {
	TimeStart string `json:"timeStart"`
	Duration  int    `json:"duration"`
	Break     int    `json:"break"`
}

// SprintsPOST saves a given sprint and returns its API location
// requires json(timeStart, duration, break)
func SprintsPOST(c *gin.Context) {
	user := c.MustGet("user").(*User)
	project := c.MustGet("project").(*Project)

	req := &SprintPOSTRequest{}
	if err := c.BindJSON(req); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	timeStart, err := time.Parse("2006-01-02T15:04:05-0700", req.TimeStart)
	if err != nil || req.Duration < 1 || req.Duration < 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	sprint, err := project.NewSprint(timeStart, req.Duration, req.Break)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header("Location", fmt.Sprintf("/users/%s/projects/%s/sprints/%s", user.Username, project.Slug, sprint.Slug))
	c.Status(http.StatusOK)
}

// SprintsSlugGET returns a specific sprint
func SprintsSlugGET(c *gin.Context) {
	sprint := c.MustGet("sprint").(*Sprint)

	c.JSON(http.StatusOK, sprint)
}

// SprintsSlugPUTRequest
type SprintsSlugPUTRequest struct {
	WordCount   int    `json:"wordCount"`
	IsMilestone bool   `json:"isMilestone"`
	Comment     string `json:"comment"`
}

// SprintsSlugPUT updates a sprint. Does not modify Slug, TimeStart or ProjectID.
// requires json(wordCount, isMilestone, comment)
func SprintsSlugPUT(c *gin.Context) {
	sprint := c.MustGet("sprint").(*Sprint)

	req := &SprintsSlugPUTRequest{}
	if err := c.BindJSON(req); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	sprint.WordCount = req.WordCount
	sprint.IsMilestone = req.IsMilestone
	sprint.Comment = req.Comment

	if err := sprint.Update(); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}
