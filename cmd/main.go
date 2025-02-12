package main

import (
	"context"
	"github.com/Quizert/PostCommentService/internal/app"
	"log"
)

func main() {
	log.Println("Starting app...")

	ctx := context.Background()
	application, err := app.InitApp(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	log.Println("Application initialized successfully")

	if err = application.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Application stopped")
}
