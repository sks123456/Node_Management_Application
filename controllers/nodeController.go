package controllers

import (
	"fmt"
	"net/http"
	"time"

	"node_management_application/config"
	"node_management_application/models"
	"node_management_application/services"
	"node_management_application/utils"

	"github.com/kataras/iris/v12"
)

// GetNodes - Fetch a list of all nodes belonging to the authenticated user
func GetNodes(ctx iris.Context) {
	// Retrieve user ID from the context
	userID := ctx.Values().GetUintDefault("user_id", 0)

	// Fetch nodes for the authenticated user
	var nodes []models.Node
	if result := config.DB.Where("user_id = ?", userID).Find(&nodes); result.Error != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": result.Error.Error()})
		return
	}

	ctx.JSON(nodes)
}

// CreateNode - Add a new node for the authenticated user
func CreateNode(ctx iris.Context) {
	// Retrieve user ID from the context
	userID := ctx.Values().GetUintDefault("user_id", 0)

	var node models.Node

	// Read request body
	if err := ctx.ReadJSON(&node); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}

	// Validate node data
	if err := services.ValidateNodeData(node.Name, node.IP, node.Port); err != nil {
		utils.ValidationErrorResponse(ctx, err)
		return
	}

	// Assign user ID and default values
	node.UserID = userID
	node.LastChecked = time.Now()
	node.Status = "Stopped"
	node.HealthStatus = "Unhealthy"

	// Save the node to the database
	if result := config.DB.Create(&node); result.Error != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to save node to database"})
		return
	}

	ctx.JSON(node)
}

// UpdateNode - Update an existing node's details belonging to the authenticated user
func UpdateNode(ctx iris.Context) {
	// Retrieve user ID from the context
	userID := ctx.Values().GetUintDefault("user_id", 0)

	var node models.Node

	// Fetch node by ID and ensure it belongs to the authenticated user
	if err := fetchNodeByIDAndUser(ctx, &node, userID); err != nil {
		return
	}

	// Read and apply updates
	var updatedData struct {
		Name     string `json:"name"`
		IP       string `json:"ip"`
		Port     int    `json:"port"`
		Location string `json:"location"`
	}
	if err := ctx.ReadJSON(&updatedData); err != nil {
		utils.ValidationErrorResponse(ctx, err)
		return
	}
	if err := services.ValidateNodeData(updatedData.Name, updatedData.IP, updatedData.Port); err != nil {
		utils.ValidationErrorResponse(ctx, err)
		return
	}

	// Apply updates
	node.Name = updatedData.Name
	node.IP = updatedData.IP
	node.Port = updatedData.Port
	node.Location = updatedData.Location

	// Save changes to the database
	if result := config.DB.Save(&node); result.Error != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update node"})
		return
	}

	ctx.JSON(node)
}

// DeleteNode - Remove a node belonging to the authenticated user
func DeleteNode(ctx iris.Context) {
	// Retrieve user ID from the context
	userID := ctx.Values().GetUintDefault("user_id", 0)

	var node models.Node

	// Fetch node by ID and ensure it belongs to the authenticated user
	if err := fetchNodeByIDAndUser(ctx, &node, userID); err != nil {
		return
	}

	// Delete the node
	if result := config.DB.Delete(&node); result.Error != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete node"})
		return
	}

	ctx.JSON(iris.Map{"message": "Node deleted successfully"})
}

// Helper: Fetch node by ID and ensure it belongs to the authenticated user
func fetchNodeByIDAndUser(ctx iris.Context, node *models.Node, userID uint) error {
	id := ctx.Params().GetUintDefault("id", 0)

	// Find the node by ID and user ID
	if result := config.DB.Where("id = ? AND user_id = ?", id, userID).First(node); result.Error != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Node not found or access denied"})
		return result.Error
	}

	return nil
}

// StartNode - Start a node belonging to the authenticated user
func StartNode(ctx iris.Context) {
	// Retrieve user ID from the context
	userID := ctx.Values().GetUintDefault("user_id", 0)

	nodeID := ctx.Params().GetUintDefault("id", 0)

	// Fetch node by ID and ensure it belongs to the authenticated user
	var node models.Node
	if result := config.DB.Where("id = ? AND user_id = ?", nodeID, userID).First(&node); result.Error != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Node not found or access denied"})
		return
	}

	// Start the node using the service
	err := services.StartNodeConcurrently(&node)
	if err != nil {
		ctx.StatusCode(http.StatusConflict)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	// Update status in the database
	if result := config.DB.Model(&node).Update("status", "Running"); result.Error != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update node status"})
		return
	}

	ctx.JSON(iris.Map{"message": "Node started successfully", "node": node})
}

// StopNode - Stop a node belonging to the authenticated user
func StopNode(ctx iris.Context) {
	// Retrieve user ID from the context
	userID := ctx.Values().GetUintDefault("user_id", 0)

	nodeID := ctx.Params().GetUintDefault("id", 0)

	// Fetch node by ID and ensure it belongs to the authenticated user
	var node models.Node
	if result := config.DB.Where("id = ? AND user_id = ?", nodeID, userID).First(&node); result.Error != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Node not found or access denied"})
		return
	}

	// Stop the node using the service
	err := services.StopNodeService(&node)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	// Update status in the database
	if result := config.DB.Model(&node).Updates(map[string]interface{}{
		"status":       "Stopped",
		"health_status": "Unhealthy",
		"last_checked": time.Now(),
	}); result.Error != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update node status"})
		return
	}

	ctx.JSON(iris.Map{"message": "Node stopped successfully", "node": node})
}

// HealthCheck - Perform a health check for a node belonging to the authenticated user
func HealthCheck(ctx iris.Context) {
	// Retrieve user ID from the context
	userID := ctx.Values().GetUintDefault("user_id", 0)

	nodeID := ctx.Params().GetUintDefault("id", 0)

	// Fetch node by ID and ensure it belongs to the authenticated user
	var node models.Node
	if result := config.DB.Where("id = ? AND user_id = ?", nodeID, userID).First(&node); result.Error != nil {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Node not found or access denied"})
		return
	}

	// Perform the health check
	err := services.PerformHealthCheckConcurrently(&node)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": fmt.Sprintf("Health check failed for node %s: %v", node.Name, err)})
		return
	}

	// Respond with the updated health status
	ctx.JSON(iris.Map{
		"node_id":       node.ID,
		"health_status": node.HealthStatus,
		"last_checked":  node.LastChecked,
	})
}

