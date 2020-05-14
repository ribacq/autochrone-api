package main

import (
	"github.com/gin-gonic/gin"

	"encoding/json"
	"fmt"
	"net/http"
)

// UsersGET sends users as JSON
func UsersGET(c *gin.Context) {
	users, err := GetUsers()
	if err != nil {
		c.JSON(http.StatusNotFound, nil)
		return
	}

	b, _ := json.Marshal(users)
	c.Data(http.StatusOK, "application/json", b)
}

// UsersPOSTRequest contains fiels for a new user
type UsersPOSTRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Confirm  string `json:"confirm"`
}

// UsersPOST registers new user
func UsersPOST(c *gin.Context) {
	req := &UsersPOSTRequest{}
	if err := c.BindJSON(req); err != nil {
		return
	}

	// TODO: check password strength
	if req.Username == "" || req.Password == "" || req.Confirm != req.Password {
		c.JSON(http.StatusUnauthorized, nil)
		return
	}

	// register user
	user, err := NewUser(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	// Respond with new user location in the api
	c.Header("Location", fmt.Sprintf("/users/%v", user.Username))
	c.JSON(http.StatusCreated, nil)
}

// UsersUsernameGET sends one user as JSON
func UsersUsernameGET(c *gin.Context) {
	user := c.MustGet("user").(*User)

	if err := user.FetchProjects(); err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, user)
}

// UsersUsernamePATCHRequest is the type for a PATCH request on /users/:username
type UsersUsernamePATCHRequest struct {
	Operator string `json:"operator"`
	Path     string `json:"path"`
	Value    string `json:"value"`
}

// UsersUsernamePATCH updates a user
func UsersUsernamePATCH(c *gin.Context) {
	user := c.MustGet("user").(*User)
	req := &UsersUsernamePATCHRequest{}
	if err := c.BindJSON(req); err != nil {
		return
	}

	// apply change
	switch req.Operator {
	case "set":
		switch req.Path {
		case "password":
			if err := user.UpdatePassword(req.Value); err != nil {
				c.JSON(http.StatusInternalServerError, nil)
				return
			}
		default:
			c.JSON(http.StatusNotFound, nil)
			return
		}
	default:
		c.JSON(http.StatusNotFound, nil)
		return
	}

	c.JSON(http.StatusOK, nil)
}

// UsersUsernameDELETE deletes a user
func UsersUsernameDELETE(c *gin.Context) {
	user := c.MustGet("user").(*User)

	// delete user
	if err := DeleteUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, nil)
}
