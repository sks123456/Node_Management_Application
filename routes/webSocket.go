package routes

import (
	"log"
	"net/http"
	websocket "node_management_application/websocket"

	WebSocket "github.com/gorilla/websocket"

	"github.com/kataras/iris/v12"
)

var upgrader = WebSocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development (update for production)
	},
}

// WebSocketHandler handles WebSocket connections
func WebSocketHandler(ctx iris.Context) {
	conn, err := upgrader.Upgrade(ctx.ResponseWriter(), ctx.Request(), nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Add the client to the websocket package's client list
	websocket.AddClient(conn)
	defer websocket.RemoveClient(conn) // Ensure client removal on disconnect

	log.Println("New WebSocket client connected")

	// Keep connection open by reading messages (or handling pings)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading WebSocket message: %v", err)
			break
		}
	}
}


// RegisterWebSocketRoute registers the WebSocket route
func RegisterWebSocketRoute(app *iris.Application) {
	app.Get("/ws", WebSocketHandler)
}
