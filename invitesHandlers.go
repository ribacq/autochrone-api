package main

import (
	"github.com/gin-gonic/gin"

	"net/http"
)

// JoinInviteSlugGET creates a guest sprint
func JoinInviteSlugGET(c *gin.Context) {
	// get current user project, and target host sprint
	project := c.MustGet("project").(*Project)
	hostSprint, err := GetSprintByInviteSlug(c.Param("islug"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// create guest sprint on user project with model host sprint
	guestSprint, err := project.NewGuestSprint(hostSprint)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, guestSprint)
}
