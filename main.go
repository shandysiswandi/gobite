package main

import (
	"context"
	"time"

	_ "github.com/go-chi/chi/v5"
	"github.com/shandysiswandi/gobite/internal/app"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	application := app.New()    // Initialize the application
	wait := application.Start() // Start the application and wait for the termination signal
	<-wait                      // Wait for the application to receive a termination signal
	application.Stop(ctx)       // Stop the application gracefully
}
