module github.com/caresle/microservices-workouts-tracker/api-gateway

go 1.25.5

require github.com/caresle/microservices-workouts-tracker/shared v0.0.0

require (
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
)

replace github.com/caresle/microservices-workouts-tracker/shared => ../shared
