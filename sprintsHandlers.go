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

// SprintsSlugPUTRequest: request for sprint update
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

// SprintsSlugNextSprintPOSTRequest: request for a next sprint
type SprintsSlugNextSprintPOSTRequest struct {
	TimeStart string `json:"timeStart"`
}

// SprintsSlugNextSprintPOST instantiates or gets the sprint following the current one.
// requires post(timeStart)
func SprintsSlugNextSprintPOST(c *gin.Context) {
	project := c.MustGet("project").(*Project)
	sprint := c.MustGet("sprint").(*Sprint)

	// singleSprint: respond with status not found
	if sprint.IsSingleSprint() {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// try to get an existing next sprint
	if nextSprint, ok := sprint.GetNextSprintIfExists(); ok {
		c.JSON(http.StatusOK, nextSprint)
		return
	}

	// get request parameter: timeStart
	req := &SprintsSlugNextSprintPOSTRequest{}
	if err := c.BindJSON(req); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	timeStart, err := time.Parse("2006-01-02T15:04:05-0700", req.TimeStart)
	if err != nil || timeStart.Before(sprint.TimeEnd()) {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// otherwise, create new sprint and return
	nextSprint, err := project.NewSprint(timeStart, sprint.Duration, sprint.Break)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, nextSprint)
}

// SprintsSlugDELETE deletes a sprint
func SprintsSlugDELETE(c *gin.Context) {
	sprint := c.MustGet("sprint").(*Sprint)

	if err := sprint.Delete(); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}
