package services

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"node_management_application/models"
)

// ServerStore to keep track of running node servers
var serverStore = sync.Map{}

// Global map to lock IP:Port combinations
var ipPortLocks = sync.Map{}

// StartNodeConcurrently starts a node with concurrency control
func StartNodeConcurrently(node *models.Node) error {
	ipPortKey := fmt.Sprintf("%s:%d", node.IP, node.Port)

	// Acquire lock for the IP:Port
	lock, _ := ipPortLocks.LoadOrStore(ipPortKey, &sync.Mutex{})
	mutex := lock.(*sync.Mutex)
	mutex.Lock()
	defer mutex.Unlock()

	// Check if the port is available
	if !isPortAvailable(node.IP, node.Port) {
		return fmt.Errorf("port %d on IP %s is already in use", node.Port, node.IP)
	}

	// Start the node server
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Node %s is running at %s:%d", node.Name, node.IP, node.Port)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", node.IP, node.Port),
		Handler: mux,
	}

	go func() {
		log.Printf("Starting node server: %s:%d", node.IP, node.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Node server %s:%d stopped with error: %v", node.IP, node.Port, err)
		}
	}()

	return nil
}
//stop Node Service
func StopNodeService(node *models.Node) error {
	value, ok := serverStore.Load(node.ID)
	if !ok {
		return fmt.Errorf("no running server found for node ID %d", node.ID)
	}

	server, ok := value.(*http.Server)
	if !ok {
		return fmt.Errorf("failed to retrieve server instance for node ID %d", node.ID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Printf("Stopping node server: %s", server.Addr)
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server for node ID %d: %v", node.ID, err)
	}

	// Remove the server from the store
	serverStore.Delete(node.ID)
	return nil
}


// isPortAvailable checks if a port is available on the given IP
func isPortAvailable(ip string, port int) bool {
	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.Listen("tcp", address)
	if err != nil {
		return false // Port is already in use
	}
	conn.Close()
	return true
}


