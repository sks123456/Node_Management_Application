package utils

import (
	"net"
	"time"
)

// Validate IP address
func IsValidIP(ip string) bool {
    return net.ParseIP(ip) != nil
}

// Check if a node is reachable
func IsNodeReachable(ip string, port int) bool {
    conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, string(port)), 3*time.Second)
    if err != nil {
        return false
    }
    conn.Close()
    return true
}
