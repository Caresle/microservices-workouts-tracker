package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/caresle/microservices-workouts-tracker/api-gateway/middleware"
	"github.com/joho/godotenv"
)

func main() {
	mux := http.NewServeMux()

	err := godotenv.Load("../.env")

	if err != nil {
		fmt.Println("Error loading .env file")
	}

	mux.HandleFunc("/api/v1/users/auth", proxyTo(os.Getenv("USER_SERVICE_URL")))

	// TODO: Add auth middleware to all routes except /api/v1/users/auth and /api/v1/users/ `POST` (for registration)

	generalMiddlewares := []func(http.Handler) http.Handler{
		middleware.RequestIDMiddleware,
		middleware.AuthMiddleware,
	}

	mux.Handle("/api/v1/users/", middleware.ChainMiddleware(proxyTo(os.Getenv("USER_SERVICE_URL")), generalMiddlewares...))
	mux.Handle("/api/v1/workouts/", middleware.ChainMiddleware(proxyTo(os.Getenv("WORKOUT_SERVICE_URL")), generalMiddlewares...))
	mux.Handle("/api/v1/exercises/", middleware.ChainMiddleware(proxyTo(os.Getenv("EXERCISE_SERVICE_URL")), generalMiddlewares...))
	mux.Handle("/api/v1/analytics/", middleware.ChainMiddleware(proxyTo(os.Getenv("ANALYTICS_SERVICE_URL")), generalMiddlewares...))

	fmt.Println("Running API Gateway....")

	http.ListenAndServe(":8080", mux)
}

func proxyTo(target string) http.HandlerFunc {
	url, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(url)

	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}
