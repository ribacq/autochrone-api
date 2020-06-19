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
	corsConfig.AllowOrigins = []string{"http://192.168.43.126:4200", "http://localhost:4200", "http://192.168.43.1:4200"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	corsConfig.AllowHeaders = []string{"Content-Type", "Authorization", "Origin"}
	corsConfig.ExposeHeaders = []string{"Location", "Access-Control-Allow-Origin"}
	r.Use(cors.New(corsConfig))

	// /auth/
	rAuth := r.Group("/auth/")
	rAuth.POST("", AuthPOST)

	// /users/
	rUsers := r.Group("/users/")
	rUsers.GET("", UsersGET)
	rUsers.POST("", UsersPOST)

	// /users/:username
	rUsersUsername := rUsers.Group("/:username")
	rUsersUsername.Use(UserLoader)
	rUsersUsername.GET("", UsersUsernameGET)
	rUsersUsername.PATCH("", TokenScopeChecker("basic"), UsersUsernamePATCH)
	rUsersUsername.DELETE("", TokenScopeChecker("basic"), UsersUsernameDELETE)

	// /users/:username/projects/
	rProjects := rUsersUsername.Group("/projects/")
	rProjects.GET("", ProjectsGET)
	rProjects.POST("", TokenScopeChecker("basic"), ProjectsPOST)

	// /users/:username/projects/:pslug
	rProjectsSlug := rProjects.Group("/:pslug")
	rProjectsSlug.Use(ProjectLoader)
	rProjectsSlug.GET("", ProjectsSlugGET)
	rProjectsSlug.PUT("", TokenScopeChecker("basic"), ProjectsSlugPUT)
	//rProjectsSlug.PATCH("", ProjectsSlugPATCH)
	rProjectsSlug.DELETE("", TokenScopeChecker("basic"), ProjectsSlugDELETE)

	// /users/:username/projects/:pslug/sprints/
	rSprints := rProjectsSlug.Group("/sprints/")
	rSprints.GET("", SprintsGET)
	rSprints.POST("", TokenScopeChecker("basic"), SprintsPOST)

	// /users/:username/projects/:pslug/sprints/:sslug
	rSprintsSlug := rSprints.Group("/:sslug")
	rSprintsSlug.Use(SprintLoader)
	rSprintsSlug.GET("", SprintsSlugGET)
	rSprintsSlug.PUT("", TokenScopeChecker("basic"), SprintsSlugPUT)
	rSprintsSlug.DELETE("", TokenScopeChecker("basic"), SprintsSlugDELETE)
	rSprintsSlug.POST("/next-sprint", TokenScopeChecker("basic"), SprintsSlugNextSprintPOST)
	rSprintsSlug.POST("/open", TokenScopeChecker("basic"), SprintsSlugOpenPOST)
	rSprintsSlug.GET("/guests", SprintsSlugGuestsGET)

	// /users/:username/projects/:pslug/join-invite/:islug
	rJoinInviteSlug := rProjectsSlug.Group("/join-invite/:islug")
	rJoinInviteSlug.GET("", TokenScopeChecker("basic"), JoinInviteSlugGET)

	r.Run(":8080")
}
