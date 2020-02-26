package router

import (
	"net/http"
	"voyager/handler/k8s"
	"voyager/handler/sd"

	_ "voyager/docs"
	"voyager/router/middleware"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

// Load loads the middlewares, routes, handler.
func Load(g *gin.Engine, mw ...gin.HandlerFunc) *gin.Engine {
	// Middlewares.
	g.Use(gin.Recovery())
	g.Use(middleware.NoCache)
	g.Use(middleware.Options)
	g.Use(middleware.Secure)
	g.Use(mw...)
	// 404 Handler.
	g.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "The incorrect API route.")
	})

	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// The health check handler
	svcdRouter := g.Group("/sd")
	{
		svcdRouter.GET("/health", sd.HealthCheck)
		svcdRouter.GET("/disk", sd.DiskCheck)
		svcdRouter.GET("/cpu", sd.CPUCheck)
		svcdRouter.GET("/ram", sd.RAMCheck)
	}

	pod := g.Group("/v1/app/pod")
	{
		//pod.POST("/create", k8s.Create)
		pod.GET("/:zone/:ns/:name", k8s.GetPod)
	}

	deployment := g.Group("/v1/app/deployment")
	{
		deployment.GET("/:zone/:ns/:name", k8s.GetDeployment)
		deployment.POST("/create", k8s.CreateDeployment)
		deployment.DELETE("/:zone/:ns/:name", k8s.DeleteDeployment)
		deployment.POST("/scale", k8s.ScaleDeployment)
		deployment.POST("/update", k8s.UpdateDeployment)
	}

	job := g.Group("/v1/app/job")
	{
		job.GET("/:zone/:ns/:name", k8s.GetJob)
		job.POST("/create", k8s.CreateJob)
		job.DELETE("/:zone/:ns/:name", k8s.DeleteJob)

	}

	return g
}
