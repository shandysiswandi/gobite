package main

import (
	"context"
	"time"

	"github.com/shandysiswandi/gobite/internal/app"
)

// @title           Gobite API
// @version         1.0
// @description     Gobite provides authentication, authorization and profile management APIs.
// @termsOfService  https://gobite.com/terms
// @contact.name    Contact Support
// @contact.url     https://gobite.com/contact
// @contact.email   support@gobite.com
// @license.name    MIT
// @license.url     https://mit-license.org/
// @server          http://localhost:8080
// @server          https://localhost:8080
// @securityDefinitions.apikey  BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT.
func main() {
	application := app.New()    // Initialize the application
	wait := application.Start() // Start the application and wait for the termination signal
	<-wait                      // Wait for the application to receive a termination signal
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	application.Stop(ctx) // Stop the application gracefully
}
