package controllers

import (
	"net/http"
	"node_management_application/config"
	"node_management_application/models"
	"node_management_application/services"
	"node_management_application/utils"
	"time"

	"github.com/kataras/iris/v12"
)

// GetNodes - Fetch a list of all nodes
func GetNodes(ctx iris.Context) {
	var nodes []models.Node

	// Fetch nodes from the database
	if result := config.DB.Find(&nodes); result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": result.Error.Error()})
		return
	}

	ctx.JSON(nodes)
}

// CreateNode - Add a new node to the database
func CreateNode(ctx iris.Context) {
	var node models.Node

	// Read request body
	if err := ctx.ReadJSON(&node); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}

	// Validate node data
	if err := services.ValidateNodeData(node.Name, node.IP, node.Port); err != nil {
		utils.ValidationErrorResponse(ctx, err)
		return
	}

	// Add default values
	node.LastChecked = time.Now()
	node.Status = "Stopped"
	node.HealthStatus = "Unhealthy"

	// Save the node to the database
	if result := config.DB.Create(&node); result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to save node to database"})
		return
	}

	ctx.JSON(node)
}


// UpdateNode - Update an existing node's details
func UpdateNode(ctx iris.Context) {
	var node models.Node

	// Fetch node by ID
	if err := fetchNodeByID(ctx, &node); err != nil {
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
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update node"})
		return
	}

	ctx.JSON(node)
}

// DeleteNode - Remove a node from the database
func DeleteNode(ctx iris.Context) {
	var node models.Node

	// Fetch node by ID
	if err := fetchNodeByID(ctx, &node); err != nil {
		return
	}

	// Delete the node
	if result := config.DB.Delete(&node); result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to delete node"})
		return
	}

	ctx.JSON(iris.Map{"message": "Node deleted successfully"})
}

// Helper: Fetch node by ID
func fetchNodeByID(ctx iris.Context, node *models.Node) error {
	id := ctx.Params().GetUintDefault("id", 0)

	// Find the node by ID
	if result := config.DB.First(node, id); result.Error != nil {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Node not found"})
		return result.Error
	}

	return nil
}

// GetUserNodes - Fetch a list of nodes belonging to a specific user by their userId
func GetUserNodes(ctx iris.Context) {
	var nodes []models.Node
	userID := ctx.Params().GetUintDefault("userId", 0)

	// Fetch nodes from the database where the UserID matches the provided userId
	if result := config.DB.Where("user_id = ?", userID).Find(&nodes); result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": result.Error.Error()})
		return
	}

	// If no nodes found
	if len(nodes) == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(iris.Map{"message": "No nodes found for this user"})
		return
	}

	ctx.JSON(nodes)
}

// StartNode starts the node by updating its status and simulating start logic
func StartNode(ctx iris.Context) {
	nodeID, err := ctx.Params().GetUint("id")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid node ID"})
		return
	}

	node, err := models.GetNodeByID(nodeID)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Node not found"})
		return
	}

	// Start the node using the service
	err = services.StartNodeConcurrently(node)
	if err != nil {
		ctx.StatusCode(http.StatusConflict)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	// Update status in the database
	err = models.UpdateNodeStatus(nodeID, "Running")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update node status"})
		return
	}

	ctx.JSON(iris.Map{"message": "Node started successfully", "node": node})
}


func StopNode(ctx iris.Context) {
	nodeID, err := ctx.Params().GetUint("id")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid node ID"})
		return
	}

	node, err := models.GetNodeByID(nodeID)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Node not found"})
		return
	}

	// Stop the node using the service
	err = services.StopNodeService(node)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to stop the node"})
		return
	}

	// Update status and health status in the database
	err = models.UpdateNodeStatusAndHealth(nodeID, "Stopped", "Unhealthy", time.Now())
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update node status and health status"})
		return
	}

	ctx.JSON(iris.Map{"message": "Node stopped successfully", "node": node})
}


func HealthCheck(ctx iris.Context) {
	nodeID, err := ctx.Params().GetUint("id")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid node ID"})
		return
	}

	node, err := models.GetNodeByID(nodeID)
	if err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": "Node not found"})
		return
	}

	// Perform health check with concurrency control
	status, err := services.PerformHealthCheckConcurrently(node)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	err = models.UpdateNodeHealthStatus(node.ID, status, time.Now())
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update health status"})
		return
	}

	ctx.JSON(iris.Map{"node_id": node.ID, "health_status": status})
}

