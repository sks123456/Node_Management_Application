package services

import (
	"fmt"
	"log"
	"net"
	"node_management_application/config"
	"node_management_application/models"
	"node_management_application/websocket"
	"sync"
	"time"
)

// Global map to lock health checks by Node ID
var healthCheckLocks = sync.Map{}

// PerformHealthCheckConcurrently checks the health of a node with concurrency control
func PerformHealthCheckConcurrently(node *models.Node) error {
	// Acquire lock for the node
	lock, _ := healthCheckLocks.LoadOrStore(node.ID, &sync.Mutex{})
	mutex := lock.(*sync.Mutex)
	mutex.Lock()
	defer func() {
		mutex.Unlock()
		healthCheckLocks.Delete(node.ID) // Clean up lock after health check
	}()

	// Perform the health check
	status, err := checkHealth(node.IP, node.Port)
	node.HealthStatus = status
	node.LastChecked = time.Now()

	// Save the updated health status to the database
	if dbErr := config.DB.Save(node).Error; dbErr != nil {
		log.Printf("Failed to update health status for node %s: %v", node.Name, dbErr)
		return fmt.Errorf("database error: %v", dbErr)
	}

	// Broadcast the health update to WebSocket clients
	websocket.BroadcastHealthStatus(node.ID, status)
	
	if err != nil {
		log.Printf("Health check failed for node %s (%s:%d): %v", node.Name, node.IP, node.Port, err)
		return fmt.Errorf("health check failed: %v", err)
	}

	log.Printf("Health check successful for node %s (%s:%d): Status is %s", node.Name, node.IP, node.Port, status)
	return nil
}

// checkHealth checks if a node is healthy by trying to connect to its port
func checkHealth(ip string, port int) (string, error) {
	address := fmt.Sprintf("%s:%d", ip, port)
	log.Printf("Performing health check on %s...", address)

	// Attempt to connect to the node's address
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		log.Printf("Health check failed for %s: %v", address, err)
		return "Unhealthy", err
	}
	defer conn.Close()

	log.Printf("Node %s is responsive", address)
	return "Healthy", nil
}
