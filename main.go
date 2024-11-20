package main

import (
	"node_management_application/config"
	"node_management_application/models"
	"node_management_application/routes"
	"node_management_application/services"

	"github.com/kataras/iris/v12"
)

func main() {
    config.ConnectDatabase()

    // Migrate the schema
    config.DB.AutoMigrate(&models.User{}, &models.Node{})
	
	 // Start the health monitoring service
	 go services.MonitorNodeHealth()
	 
	app := iris.New()
    routes.RegisterRoutes(app)

    app.Listen(":8080")
}
