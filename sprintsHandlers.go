package main

import (
	"github.com/gin-gonic/gin"

	"net/http"
)

// SprintsGET responds with sprints according to the given filters (username&projectSlug, or sprintSlug)
func SprintsGET(c *gin.Context) {
	if sprintSlug, ok := c.GetQuery("sprintSlug"); ok {
		sprint, err := GetSprintBySlug(sprintSlug)
		if err != nil {
			c.JSON(http.StatusNotFound, nil)
			return
		}
		c.JSON(http.StatusOK, sprint)
		return
	}

	username, okUsername := c.GetQuery("username")
	projectSlug, okProjectSlug := c.GetQuery("projectSlug")
	byDate, okByDate := c.GetQuery("byDate")

	if !okUsername || !okProjectSlug {
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	user, err := GetUserByUsername(username)
	if err != nil {
		c.JSON(http.StatusNotFound, nil)
		return
	}

	project, err := user.GetProjectBySlug(projectSlug)
	if err != nil {
		c.JSON(http.StatusNotFound, nil)
		return
	}

	err = project.FetchSprints()
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	if okByDate && byDate == "true" {
		sprintsByDate, err := project.GetSprintsByDate()
		if err != nil {
			c.JSON(http.StatusInternalServerError, nil)
			return
		}

		c.JSON(http.StatusOK, sprintsByDate)
		return
	}

	c.JSON(http.StatusOK, project.Sprints)
}
