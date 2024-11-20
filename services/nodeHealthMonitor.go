package services

import (
	"net"
	"node_management_application/config"
	"node_management_application/models"
	"time"
)

func MonitorNodeHealth() {
    ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
    defer ticker.Stop()

    for {
        <-ticker.C
        var nodes []models.Node
        config.DB.Find(&nodes)

        for _, node := range nodes {
            if isNodeReachable(node.IP, node.Port) {
                node.HealthStatus = "Healthy"
            } else {
                node.HealthStatus = "Unhealthy"
            }
            config.DB.Save(&node)
        }
    }
}

func isNodeReachable(ip string, port int) bool {
    conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, string(port)), 3*time.Second)
    if err != nil {
        return false
    }
    conn.Close()
    return true
}
