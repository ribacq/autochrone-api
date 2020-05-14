package main

import (
	"github.com/gin-gonic/gin"

	"net/http"
)

// AuthGETRequest contains authentication fields
type AuthGETRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthGET replies to an authentication request with a JSON token or error message
// expects get(username, password)
func AuthGET(c *gin.Context) {
	// gets username and password
	req := &AuthGETRequest{}
	if err := c.BindJSON(req); err != nil {
		return
	}

	// get user or replies to request with an error
	user, err := GetUserByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// authenticates user or replies with an error
	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}

	// generate token, maybe reply with an error
	token, err := user.GenerateToken("basic")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	// OK: reply with token
	c.JSON(http.StatusOK, gin.H{"token": token})
}
