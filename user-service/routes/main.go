package routes

import (
	"time"

	"github.com/caresle/microservices-workouts-tracker/shared"
	"github.com/gin-gonic/gin"
)

var router = gin.Default()

func Run() {
	getRoutes()

	router.Run(":8081")
}

func getRoutes() {
	v1 := router.Group("/api/v1")

	v1.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, shared.ApiResponse{
			Data:      []any{"User service tracker"},
			Timestamp: time.Now().Unix(),
		})
	})
}
