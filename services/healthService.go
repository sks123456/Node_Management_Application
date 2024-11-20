package services

import (
	"fmt"
	"log"
	"net"
	"node_management_application/models"
	"sync"
	"time"
)

// Global map to lock health checks by Node ID
var healthCheckLocks = sync.Map{}

// PerformHealthCheckConcurrently checks the health of a node with concurrency control
func PerformHealthCheckConcurrently(node *models.Node) (string, error) {
	// Acquire lock for the node
	lock, _ := healthCheckLocks.LoadOrStore(node.ID, &sync.Mutex{})
	mutex := lock.(*sync.Mutex)
	mutex.Lock()
	defer mutex.Unlock()

	// Perform the health check
	status, err := checkHealth(node.IP, node.Port)
	return status, err
}

// checkHealth checks if a node is healthy by trying to connect to its port
func checkHealth(ip string, port int) (string, error) {
	address := fmt.Sprintf("%s:%d", ip, port)
	log.Printf("Checking health of %s", address)

	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		log.Printf("Health check failed for %s: %v", address, err)
		return "Unhealthy", err
	}
	defer conn.Close()

	log.Printf("Port %s is responsive", address)
	return "Healthy", nil
}
