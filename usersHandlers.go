package main

import (
	"github.com/gin-gonic/gin"

	"encoding/json"
	"fmt"
	"log"
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

// UsersPOST registers new user
// expects post(username, password, confirm)
func UsersPOST(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	confirm := c.PostForm("confirm")

	if username == "" || password == "" || confirm != password {
		c.JSON(http.StatusUnauthorized, nil)
		return
	}

	user, err := NewUser(username, password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.Header("Location", fmt.Sprintf("/users/%v", user.ID))
	c.JSON(http.StatusCreated, nil)
}

// UsersUsernameGET sends one user as JSON
// expects param(username)
func UsersUsernameGET(c *gin.Context) {
	user, err := GetUserByUsername(c.Param("username"))
	if err != nil {
		c.JSON(http.StatusNotFound, nil)
		return
	}

	if err := user.FetchProjects(); err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, user)
}

// UsersUsernamePATCH updates a user
// expects param(username), post(operator, path, value)
func UsersUsernamePATCH(c *gin.Context) {
	user, err := GetUserByUsername(c.Param("username"))
	if err != nil {
		c.JSON(http.StatusNotFound, nil)
		return
	}

	log.Print(c.PostForm("operator"))

	switch c.PostForm("operator") {
	case "set":
		switch c.PostForm("path") {
		case "password":
			if err := user.UpdatePassword(c.PostForm("value")); err != nil {
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
// expects param(username)
func UsersUsernameDELETE(c *gin.Context) {
	user, err := GetUserByUsername(c.Param("username"))
	if err != nil {
		c.JSON(http.StatusNotFound, nil)
		return
	}

	if err := DeleteUser(user); err != nil {
		log.Print(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, nil)
}
