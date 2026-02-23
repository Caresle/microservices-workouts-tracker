package routes

import (
	"strconv"
	"time"

	"github.com/caresle/microservices-workouts-tracker/shared"
	"github.com/caresle/microservices-workouts-tracker/user-service/lib"
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

	v1.POST("/auth", func(ctx *gin.Context) {
		var body user_request.AuthUserRequest

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

		user, err := queries.ValidateUserCredentials(body.Email, body.Password)

		if err != nil {
			ctx.JSON(500, shared.ApiResponse{
				Errors: []shared.APIError{
					{
						Code:    "INTERNAL_SERVER_ERROR",
						Message: "Failed to authenticate user",
						Details: map[string]interface{}{
							"error": err.Error(),
						},
					},
				},
				Timestamp: time.Now().Unix(),
			})
			return
		}

		if user == nil {
			ctx.JSON(401, shared.ApiResponse{
				Errors: []shared.APIError{
					{
						Code:    "UNAUTHORIZED",
						Message: "Invalid email or password",
					},
				},
				Timestamp: time.Now().Unix(),
			})
			return
		}

		token, err := lib.GenerateJWT(user)

		if err != nil {
			ctx.JSON(500, shared.ApiResponse{
				Errors: []shared.APIError{
					{
						Code:    "INTERNAL_SERVER_ERROR",
						Message: "Failed to generate JWT",
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
			Data:      []any{map[string]string{"token": token}},
			Timestamp: time.Now().Unix(),
		})
	})

	v1.PUT("/", func(ctx *gin.Context) {
		var body user_request.UpdateUserRequest

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

		userBody, password := models.FromUpdateRequestToUser(body)
		user, err := queries.UpdateUser(*userBody, password)

		if err != nil {
			ctx.JSON(500, shared.ApiResponse{
				Errors: []shared.APIError{
					{
						Code:    "INTERNAL_SERVER_ERROR",
						Message: "Failed to update user",
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

	v1.DELETE("/:id", func(ctx *gin.Context) {
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

		err = queries.DeleteUser(id)

		if err != nil {
			ctx.JSON(500, shared.ApiResponse{
				Errors: []shared.APIError{
					{
						Code:    "INTERNAL_SERVER_ERROR",
						Message: "Failed to delete user",
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
			Data:      []any{map[string]int{"deleted_user_id": id}},
			Timestamp: time.Now().Unix(),
		})
	})
}
