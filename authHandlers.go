package main

import (
	"github.com/gin-gonic/gin"

	"net/http"
)

// AuthPOSTRequest contains authentication fields
type AuthPOSTRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthPOST replies to an authentication request with a JSON token or error message
func AuthPOST(c *gin.Context) {
	// gets username and password
	req := &AuthPOSTRequest{}
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
