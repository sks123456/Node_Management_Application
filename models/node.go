package models

import (
	"node_management_application/config"
	"time"

	"gorm.io/gorm"
)

var DB *gorm.DB // Assume this is initialized elsewhere

type Node struct {
	ID           uint      `gorm:"primaryKey"`
	UserID       uint      `gorm:"not null"`
	Name         string    `gorm:"size:100;not null"`
	IP           string    `gorm:"size:50;not null"`
	Status       string    `gorm:"size:50;default:'Stopped'"`
	HealthStatus string    `gorm:"size:50;default:'Healthy'"`
	Location     string    `gorm:"size:100"`
	Port         int
	LastChecked  time.Time `gorm:"autoCreateTime"`
}

func GetNodeByID(id uint) (*Node, error) {
	var node Node
	if err := config.DB.First(&node, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}
	return &node, nil
}

func UpdateNodeStatus(id uint, status string) error {
	return config.DB.Model(&Node{}).Where("id = ?", id).Update("status", status).Error
}

func UpdateNodeHealthStatus(id uint, healthStatus string, lastChecked time.Time) error {
	return config.DB.Model(&Node{}).Where("id = ?", id).Updates(map[string]interface{}{
		"health_status": healthStatus,
		"last_checked":  lastChecked,
	}).Error
}

func UpdateNodeStatusAndHealth(id uint, status string, healthStatus string, lastChecked time.Time) error {
	return config.DB.Model(&Node{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        status,
		"health_status": healthStatus,
		"last_checked":  lastChecked,
	}).Error
}