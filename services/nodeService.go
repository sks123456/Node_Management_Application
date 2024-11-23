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

	// Create an HTTP server for the node
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Node %s is running at %s:%d", node.Name, node.IP, node.Port)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", node.IP, node.Port),
		Handler: mux,
	}

	// Track the server in the store
	serverStore.Store(node.ID, server)

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting node server: %s:%d", node.IP, node.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Node server %s:%d stopped with error: %v", node.IP, node.Port, err)
			// Clean up on failure
			serverStore.Delete(node.ID)
		}
	}()

	return nil
}

// StopNodeService stops a running node server
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

	// Remove the server from the store and release the IP:Port lock
	serverStore.Delete(node.ID)
	ipPortLocks.Delete(fmt.Sprintf("%s:%d", node.IP, node.Port))
	return nil
}

// isPortAvailable checks if a port is available on the given IP
func isPortAvailable(ip string, port int) bool {
	address := fmt.Sprintf("%s:%d", ip, port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Printf("Port check failed for %s: %v", address, err)
		return false // Port is already in use or permission denied
	}
	defer listener.Close()
	return true
}

func StopAllNodes() {
	serverStore.Range(func(key, value interface{}) bool {
		server := value.(*http.Server)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Failed to shutdown server %s: %v", server.Addr, err)
		} else {
			log.Printf("Server %s stopped successfully", server.Addr)
		}
		serverStore.Delete(key)
		return true
	})
}
