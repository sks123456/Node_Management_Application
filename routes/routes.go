package routes

import (
	"node_management_application/controllers"

	"github.com/kataras/iris/v12"
)

func RegisterRoutes(app *iris.Application) {
	 // User routes
     userAPI := app.Party("/users")
     {
        userAPI.Get("/", controllers.GetUsers)
        userAPI.Post("/", controllers.CreateUser)
        userAPI.Put("/{id:uint}", controllers.UpdateUser)   // Update user
        userAPI.Delete("/{id:uint}", controllers.DeleteUser) // Delete user
        userAPI.Get("/{userId:uint}/nodes", controllers.GetUserNodes) // Get User nodes

     }
 
     // Node routes
     nodeAPI := app.Party("/nodes")
     {
        nodeAPI.Get("/", controllers.GetNodes)
        nodeAPI.Post("/", controllers.CreateNode)
         nodeAPI.Put("/{id:uint}", controllers.UpdateNode)
         nodeAPI.Delete("/{id:uint}", controllers.DeleteNode)
         nodeAPI.Post("/{id:uint}/start", controllers.StartNode)
         nodeAPI.Post("/{id:uint}/stop", controllers.StopNode)
         nodeAPI.Get("/{id:uint}/health", controllers.HealthCheck)
     }
     
    app.Get("/ping", func(ctx iris.Context) {
        ctx.WriteString("pong")
    })
}
