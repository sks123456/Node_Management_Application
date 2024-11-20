package services

import (
	"errors"
	"net"
)

// ValidateNodeData validates the node data for required fields and proper formatting
func ValidateNodeData(name string, ip string, port int) error {
	if name == "" {
		return errors.New("name is required")
	}
	if ip == "" {
		return errors.New("IP address is required")
	}
	if net.ParseIP(ip) == nil {
		return errors.New("invalid IP address format")
	}
	if port <= 0 || port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	return nil
}
