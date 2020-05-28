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
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	corsConfig.ExposeHeaders = []string{"Access-Control-Allow-Origin", "Location"}
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
	rProjects := rUsersUsername.Group("/projects")
	rProjects.GET("/", ProjectsGET)
	rProjects.POST("/", TokenScopeChecker("basic"), ProjectsPOST)

	// /users/:username/projects/:pslug
	rProjectsSlug := rProjects.Group("/:pslug")
	rProjectsSlug.Use(ProjectLoader)
	rProjectsSlug.GET("", ProjectsSlugGET)
	rProjectsSlug.PUT("", TokenScopeChecker("basic"), ProjectsSlugPUT)
	//rProjectsSlug.PATCH("", ProjectsSlugPATCH)
	rProjectsSlug.DELETE("", TokenScopeChecker("basic"), ProjectsSlugDELETE)

	// /users/:username/projects/:pslug/sprints/
	rSprints := rProjectsSlug.Group("/sprints")
	rSprints.GET("/", SprintsGET)
	rSprints.POST("/", TokenScopeChecker("basic"), SprintsPOST)

	// /users/:username/sprints/:sslug
	/*rSprints.GET("/:sslug", SprintsSlugGET)
	rSprints.PATCH("/:sslug", SprintsSlugPATCH)
	rSprints.DELETE("/:sslug", SprintsSlugDELETE)*/

	r.Run(":8080")
}
