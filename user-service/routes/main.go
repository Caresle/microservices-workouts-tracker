package routes

import "github.com/gin-gonic/gin"

var router = gin.Default()

func Run() {
	getRoutes()

	router.Run(":8081")
}

func getRoutes() {
	v1 := router.Group("/api/v1")

	v1.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "User service tracker",
		})
	})
}
