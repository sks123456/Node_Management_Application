package services

import (
	"log"
	"node_management_application/config"
	"node_management_application/models"
	"time"
)

// MonitorNodeHealth periodically checks the health of all nodes
func MonitorNodeHealth(shutdown chan struct{}) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-shutdown:
			log.Println("Health monitoring service shutting down...")
			return
		case <-ticker.C:
			log.Println("Performing health checks for all nodes...")
			var nodes []models.Node
			config.DB.Where("status = ?","Running").Find(&nodes)
			for _, node := range nodes {
				go func(n models.Node) {
					if err := PerformHealthCheckConcurrently(&n); err != nil {
						log.Printf("Health check error for node %s: %v", n.Name, err)
					}
				}(node)
			}
		}
	}
}
