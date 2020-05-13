// autochrone-api
package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	// gin router
	r := gin.Default()

	// auth
	rAuth := r.Group("/auth")
	rAuth.POST("/", AuthPOST)

	// users
	rUsers := r.Group("/users")
	rUsers.GET("/", UsersGET)
	rUsers.POST("/", UsersPOST)
	rUsers.GET("/:username", UsersUsernameGET)
	rUsers.PATCH("/:username", UsersUsernamePATCH)
	rUsers.DELETE("/:username", UsersUsernameDELETE)

	/*/ projects
	rProjects := rUsers.Group("/:username/projects")
	rProjects.GET("/", ProjectsGET)
	rProjects.POST("/", ProjectsPOST)
	rProjects.GET("/:slug", ProjectsIdGET)
	rProjects.PATCH("/:slug", ProjectsIdPATCH)
	rProjects.DELETE("/:slug", ProjectsIdDELETE)

	// sprints
	rSprints := r.Group("/sprints")
	rSprints.GET("/", SprintsGET)
	rSprints.POST("/", SprintsPOST)
	rSprints.GET("/:id", SprintsIdGET)
	rSprints.PATCH("/:id", SprintsIdPATCH)
	rSprints.DELETE("/:id", SprintsIdDELETE)*/

	r.Run(":8080")
}
