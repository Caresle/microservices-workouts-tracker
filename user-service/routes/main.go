package routes

import (
	"strconv"
	"time"

	"github.com/caresle/microservices-workouts-tracker/shared"
	"github.com/caresle/microservices-workouts-tracker/user-service/models"
	"github.com/caresle/microservices-workouts-tracker/user-service/queries"
	user_request "github.com/caresle/microservices-workouts-tracker/user-service/request"
	"github.com/gin-gonic/gin"
)

var router = gin.Default()

func Run() {
	getRoutes()

	router.Run(":8081")
}

func getRoutes() {
	v1 := router.Group("/api/v1/users")

	v1.GET("/", func(ctx *gin.Context) {
		users, err := queries.GetAllUsers()

		if err != nil {
			ctx.JSON(500, shared.ApiResponse{
				Errors: []shared.APIError{
					{
						Code:    "INTERNAL_SERVER_ERROR",
						Message: "Failed to get users",
						Details: map[string]interface{}{
							"error": err.Error(),
						},
					},
				},
				Timestamp: time.Now().Unix(),
			})
			return
		}

		ctx.JSON(200, shared.ApiResponse{
			Data:      []any{users},
			Timestamp: time.Now().Unix(),
		})
	})

	v1.GET("/:id", func(ctx *gin.Context) {
		id, err := strconv.Atoi(ctx.Param("id"))

		if err != nil {
			ctx.JSON(400, shared.ApiResponse{
				Errors: []shared.APIError{
					{
						Code:    "VALIDATION_ERROR",
						Message: "Invalid user id",
						Field:   "id",
						Details: map[string]interface{}{
							"error": err.Error(),
						},
					},
				},
				Timestamp: time.Now().Unix(),
			})
			return
		}

		user, err := queries.GetUserById(id)

		if err != nil {
			ctx.JSON(500, shared.ApiResponse{
				Errors: []shared.APIError{
					{
						Code:    "INTERNAL_SERVER_ERROR",
						Message: "Failed to get user",
						Details: map[string]interface{}{
							"error": err.Error(),
						},
					},
				},
				Timestamp: time.Now().Unix(),
			})
			return
		}

		ctx.JSON(200, shared.ApiResponse{
			Data:      []any{user},
			Timestamp: time.Now().Unix(),
		})
	})

	v1.POST("/", func(ctx *gin.Context) {
		var body user_request.CreateUserRequest

		if err := ctx.ShouldBindJSON(&body); err != nil {
			ctx.JSON(400, shared.ApiResponse{
				Errors: []shared.APIError{
					{
						Code:    "VALIDATION_ERROR",
						Message: "Invalid request body",
						Field:   "body",
						Details: map[string]interface{}{
							"error": err.Error(),
						},
					},
				},
				Timestamp: time.Now().Unix(),
			})
			return
		}

		userBody, password := models.FromCreateRequestToUser(body)
		user, err := queries.CreateUser(*userBody, password)

		if err != nil {
			ctx.JSON(500, shared.ApiResponse{
				Errors: []shared.APIError{
					{
						Code:    "INTERNAL_SERVER_ERROR",
						Message: "Failed to create user",
						Details: map[string]interface{}{
							"error": err.Error(),
						},
					},
				},
				Timestamp: time.Now().Unix(),
			})
			return
		}

		ctx.JSON(200, shared.ApiResponse{
			Data:      []any{*user},
			Timestamp: time.Now().Unix(),
		})
	})
}
