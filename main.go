package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"node_management_application/config"
	"node_management_application/models"
	"node_management_application/routes"
	"node_management_application/services"

	"github.com/kataras/iris/v12"
)

func main() {
	// Initialize the application
	initialize()

	// Start the health monitoring service
	healthMonitorShutdown := startHealthMonitoring()

	// Start the Iris web server
	app := startServer()

	// Handle graceful shutdown
	handleShutdown(app, healthMonitorShutdown)
}

// initialize sets up the database and performs migrations
func initialize() {
	log.Println("Initializing application...")

	// Connect to the database
	log.Println("Connecting to the database...")
	config.ConnectDatabase()

	// Run database migrations
	log.Println("Running database migrations...")
	if err := config.DB.AutoMigrate(&models.User{}, &models.Node{}); err != nil {
		log.Fatalf("Failed to migrate database schema: %v", err)
	}
}

// startHealthMonitoring starts the health monitoring service in a goroutine
func startHealthMonitoring() chan struct{} {
	log.Println("Starting health monitoring service...")
	shutdown := make(chan struct{})
	go services.MonitorNodeHealth(shutdown)
	return shutdown
}

// startServer initializes and starts the Iris web server
func startServer() *iris.Application {
	log.Println("Starting web server...")

	app := iris.New()

	// Register routes
	log.Println("Registering routes...")
	routes.RegisterRoutes(app)

	// Start the server in a goroutine
	go func() {
		if err := app.Listen(":8080"); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	return app
}

// handleShutdown manages the cleanup and graceful shutdown of the application
func handleShutdown(app *iris.Application, healthMonitorShutdown chan struct{}) {
	// Set up signal channel to catch OS interrupts
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Wait for a termination signal
	<-shutdown
	log.Println("Shutting down application...")

	// Stop health monitoring service
	log.Println("Stopping health monitoring service...")
	close(healthMonitorShutdown)

	// Stop all node servers
	log.Println("Stopping all node servers...")
	services.StopAllNodes()

	// Close database connections
	log.Println("Closing database connections...")
	sqlDB, err := config.DB.DB() // Access the underlying *sql.DB
	if err != nil {
		log.Printf("Failed to access database connection: %v", err)
	} else {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Failed to close database connection: %v", err)
		}
	}

	// Shutdown the web server
	log.Println("Shutting down web server...")
	if err := app.Shutdown(nil); err != nil {
		log.Printf("Failed to shutdown web server: %v", err)
	}

	log.Println("Application shutdown completed.")
}
