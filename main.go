// autochrone-api
package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// gin router
	r := gin.Default()

	// CORS settings
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:4200"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PATCH", "DELETE"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type"}
	corsConfig.ExposeHeaders = []string{"Access-Control-Allow-Origin"}
	r.Use(cors.New(corsConfig))

	// /auth/
	rAuth := r.Group("/auth")
	rAuth.POST("/", AuthPOST)

	// /users/
	rUsers := r.Group("/users")
	rUsers.GET("/", UsersGET)
	rUsers.POST("/", UsersPOST)

	// /users/:username
	rUsersUsername := rUsers.Group("/:username")
	rUsersUsername.Use(UserLoader)
	rUsersUsername.GET("", UsersUsernameGET)
	rUsersUsername.PATCH("", TokenScopeChecker("basic"), UsersUsernamePATCH)
	rUsersUsername.DELETE("", TokenScopeChecker("basic"), UsersUsernameDELETE)

	// /users/:username/projects/
	rProjects := rUsers.Group("/:username/projects")
	rProjects.Use(UserLoader)
	rProjects.GET("/", ProjectsGET)
	rProjects.POST("/", TokenScopeChecker("basic"), ProjectsPOST)

	// /users/:username/projects/:slug
	rProjectsSlug := rProjects.Group("/:slug")
	rProjectsSlug.Use(ProjectLoader)
	rProjectsSlug.GET("", ProjectsSlugGET)
	rProjectsSlug.PUT("", TokenScopeChecker("basic"), ProjectsSlugPUT)
	//rProjectsSlug.PATCH("", ProjectsSlugPATCH)
	rProjectsSlug.DELETE("", TokenScopeChecker("basic"), ProjectsSlugDELETE)

	// /sprints/
	rSprints := r.Group("/sprints")
	rSprints.GET("/", SprintsGET)
	/*rSprints.POST("/", SprintsPOST)
	rSprints.GET("/:id", SprintsIdGET)
	rSprints.PATCH("/:id", SprintsIdPATCH)
	rSprints.DELETE("/:id", SprintsIdDELETE)*/

	r.Run(":8080")
}
