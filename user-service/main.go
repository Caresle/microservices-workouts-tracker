package main

import (
	"fmt"

	"github.com/caresle/microservices-workouts-tracker/user-service/routes"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		fmt.Println("Error loading .env file")
	}

	fmt.Println("Running user service....")
	routes.Run()
}
