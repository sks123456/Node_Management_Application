package websocket

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	clients    = make(map[*websocket.Conn]bool) // Connected WebSocket clients
	clientsMux = sync.Mutex{}                   // Mutex for thread-safe operations
)

// AddClient adds a new WebSocket client
func AddClient(conn *websocket.Conn) {
	clientsMux.Lock()
	clients[conn] = true
	clientsMux.Unlock()
	log.Println("New WebSocket client added")
}

// RemoveClient removes a WebSocket client
func RemoveClient(conn *websocket.Conn) {
	clientsMux.Lock()
	delete(clients, conn)
	clientsMux.Unlock()
	conn.Close()
	log.Println("WebSocket client removed")
}

// BroadcastHealthStatus sends a health update to all connected WebSocket clients
func BroadcastHealthStatus(nodeID uint, healthStatus string) {
    log.Printf("Broadcasting health status: node_id=%d, health_status=%s", nodeID, healthStatus)

    message := map[string]interface{}{
		"node_id":       nodeID,
		"health_status": healthStatus,
	}

	clientsMux.Lock()
	defer clientsMux.Unlock()

	for client := range clients {
		err := client.WriteJSON(message)
		if err != nil {
			log.Printf("Failed to send message to WebSocket client: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}
