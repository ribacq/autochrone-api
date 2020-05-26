package main

import (
	"github.com/gin-gonic/gin"

	"fmt"
	"log"
	"net/http"
	"time"
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

// ProjectRequest determines fields for a project request
type ProjectRequest struct {
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	DateStart      string `json:"dateStart"`
	DateEnd        string `json:"dateEnd"`
	WordCountStart int    `json:"wordCountStart"`
	WordCountGoal  int    `json:"wordCountGoal"`
}

// ProjectsPOST adds a new project and responds with its API location in a Location header
func ProjectsPOST(c *gin.Context) {
	user := c.MustGet("user").(*User)
	req := &ProjectRequest{}
	if err := c.BindJSON(req); err != nil {
		return
	}

	dateStart, errDateStart := time.Parse("2006-01-02", req.DateStart)
	dateEnd, errDateEnd := time.Parse("2006-01-02", req.DateEnd)
	if errDateStart != nil || errDateEnd != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	project, err := user.NewProject(req.Name, req.Slug, dateStart, dateEnd, req.WordCountStart, req.WordCountGoal)
	if err != nil {
		log.Print(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header("Location", fmt.Sprintf("/users/%s/projects/%s", user.Username, project.Slug))
	c.Status(http.StatusOK)
}

// ProjectsSlugGET responds with a single project for a given user
func ProjectsSlugGET(c *gin.Context) {
	project := c.MustGet("project").(*Project)

	c.JSON(http.StatusOK, project)
}

// ProjectsSlugPUT updates a whole project
func ProjectsSlugPUT(c *gin.Context) {
	project := c.MustGet("project").(*Project)
	req := &ProjectRequest{}
	if err := c.BindJSON(req); err != nil {
		return
	}

	if project.Slug != req.Slug {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	dateStart, errDateStart := time.Parse("2006-01-02", req.DateStart)
	dateEnd, errDateEnd := time.Parse("2006-01-02", req.DateEnd)
	if errDateStart != nil || errDateEnd != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	project.Name = req.Name
	project.DateStart = dateStart
	project.DateEnd = dateEnd
	project.WordCountStart = req.WordCountStart
	project.WordCountGoal = req.WordCountGoal
	if err := project.Update(); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

// ProjectsSlugDELETE deletes a whole project and all its sprints
func ProjectsSlugDELETE(c *gin.Context) {
	project := c.MustGet("project").(*Project)

	if err := project.Delete(); err != nil {
		c.Status(http.StatusInternalServerError)
	}

	c.Status(http.StatusOK)
}
