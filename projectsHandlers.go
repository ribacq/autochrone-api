package main

import (
	"github.com/gin-gonic/gin"

	"net/http"
)

// ProjectsGET responds with all projects for a given user
func ProjectsGET(c *gin.Context) {
	user := c.MustGet("user").(*User)
	if err := user.FetchProjects(); err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, user.Projects)
}

// ProjectsSlugGET responds with a single project for a given user
func ProjectsSlugGET(c *gin.Context) {
	project := c.MustGet("project").(*Project)

	c.JSON(http.StatusOK, project)
}
