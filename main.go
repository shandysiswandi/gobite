package main

import (
	"context"
	"time"

	"github.com/shandysiswandi/gobite/internal/app"
)

// @title           Gobite API
// @version         1.0
// @description     Gobite provides authentication and profile management APIs.
// @termsOfService  https://gobite.com/terms
// @contact.name    Gobite API Support
// @contact.url     https://gobite.com/contact
// @contact.email   support@gobite.com
// @license.name    MIT
// @license.url     https://mit-license.org/
// @host            localhost:8080
// @BasePath        /
// @schemes         http https
// @securityDefinitions.bearerauth  BearerAuth
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	application := app.New()    // Initialize the application
	wait := application.Start() // Start the application and wait for the termination signal
	<-wait                      // Wait for the application to receive a termination signal
	application.Stop(ctx)       // Stop the application gracefully
}
