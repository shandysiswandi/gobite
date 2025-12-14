package main

import (
	"context"
	"time"

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
