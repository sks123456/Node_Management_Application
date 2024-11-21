package routes

import (
	"node_management_application/controllers"
	"node_management_application/middlewares"

	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
)

func RegisterRoutes(app *iris.Application) {
    // Enable CORS middleware
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8081"}, // Replace with the frontend's origin
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	app.UseRouter(corsMiddleware)

    // Authentication routes
    app.Post("/register", controllers.RegisterUser)
	app.Post("/login", controllers.Login)


	 // User routes
     userAPI := app.Party("/users")
     {
        userAPI.Get("/", controllers.GetUsers)
        // userAPI.Post("/", controllers.CreateUser)
        userAPI.Put("/{id:uint}", controllers.UpdateUser)   // Update user
        userAPI.Delete("/{id:uint}", controllers.DeleteUser) // Delete user
        userAPI.Get("/{userId:uint}/nodes", controllers.GetUserNodes) // Get User nodes

     }
 
     // Node routes
     nodeAPI := app.Party("/nodes",middlewares.Authenticate)
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
