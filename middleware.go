package main

import (
	"github.com/gin-gonic/gin"

	"fmt"
	"net/http"
	"strings"
)

// UserLoader: middleware that sets context user using request param :username
func UserLoader(c *gin.Context) {
	user, err := GetUserByUsername(c.Param("username"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("user not found %q", c.Param("username"))})
		return
	}

	c.Set("user", user)
}

// ProjectLoader: middleware that sets context project using request param :pslug
// Must be used after UserLoader
func ProjectLoader(c *gin.Context) {
	user := c.MustGet("user").(*User)
	project, err := user.GetProjectBySlug(c.Param("pslug"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("project not found %q for user %q", c.Param("pslug"), user.Username)})
		return
	}

	c.Set("project", project)
}

// SprintLoader: middleware that sets context sprint using request param :sslug
func SprintLoader(c *gin.Context) {
	sprint, err := GetSprintBySlug(c.Param("sslug"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("sprint not found %q", c.Param("sslug"))})
		return
	}

	c.Set("sprint", sprint)
}

// TokenScopeChecker: returns a middleware that checks for a given scope.
// Requires UserLoader middleware to have been called first.
func TokenScopeChecker(scope string) func(*gin.Context) {
	return func(c *gin.Context) {
		user := c.MustGet("user").(*User)

		tokenString := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		ok, err := user.TokenValidInScope(tokenString, scope)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("invalid token error: %v", err)})
			return
		} else if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("invalid token for scope %q", scope)})
			return
		}
	}
}
