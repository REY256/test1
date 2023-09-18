package main

import (
	"log"
	"test1/internal/pkg/app"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

func main() {
	app, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	app.Run()
}
