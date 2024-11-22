package controllers

import (
	"node_management_application/config"
	"node_management_application/models"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

func GetUsers(ctx iris.Context) {
	var users []models.User
	result := config.DB.Find(&users)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": result.Error.Error()})
		return
	}
	ctx.JSON(users)
}

func GetUserProfile(ctx iris.Context) {
    // Get the user ID from the context (assumes it's set by authentication middleware)
    userID := ctx.Values().GetUintDefault("user_id", 0)

    if userID == 0 {
        ctx.StatusCode(iris.StatusBadRequest)
        ctx.JSON(iris.Map{"error": "User ID is missing"})
        return
    }

    var user models.User

    // Find the user by userID
    result := config.DB.First(&user, userID) // This fetches the first matching user
    if result.Error != nil {
        if result.Error == gorm.ErrRecordNotFound {
            ctx.StatusCode(iris.StatusNotFound)
            ctx.JSON(iris.Map{"error": "User not found"})
        } else {
            ctx.StatusCode(iris.StatusInternalServerError)
            ctx.JSON(iris.Map{"error": result.Error.Error()})
        }
        return
    }

    // Return the user details in the response
    ctx.JSON(user)
}


func CreateUser(ctx iris.Context) {
	var user models.User
	if err := ctx.ReadJSON(&user); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}
	result := config.DB.Create(&user)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": result.Error.Error()})
		return
	}
	ctx.JSON(user)
}

func UpdateUser(ctx iris.Context) {
    id := ctx.Params().GetUintDefault("id", 0)
    var user models.User

    // Find the user by ID
    if result := config.DB.First(&user, id); result.Error != nil {
        ctx.StatusCode(iris.StatusNotFound)
        ctx.JSON(iris.Map{"error": "User not found"})
        return
    }

    // Read and apply updates
    var updatedData struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }
    if err := ctx.ReadJSON(&updatedData); err != nil {
        ctx.StatusCode(iris.StatusBadRequest)
        ctx.JSON(iris.Map{"error": err.Error()})
        return
    }

    user.Name = updatedData.Name
    user.Email = updatedData.Email

    // Save changes to the database
    if result := config.DB.Save(&user); result.Error != nil {
        ctx.StatusCode(iris.StatusInternalServerError)
        ctx.JSON(iris.Map{"error": result.Error.Error()})
        return
    }

    ctx.JSON(user)
}

func DeleteUser(ctx iris.Context) {
    id := ctx.Params().GetUintDefault("id", 0)
    var user models.User

    // Find the user by ID
    if result := config.DB.First(&user, id); result.Error != nil {
        ctx.StatusCode(iris.StatusNotFound)
        ctx.JSON(iris.Map{"error": "User not found"})
        return
    }

    // Delete the user
    if result := config.DB.Delete(&user); result.Error != nil {
        ctx.StatusCode(iris.StatusInternalServerError)
        ctx.JSON(iris.Map{"error": result.Error.Error()})
        return
    }

    ctx.JSON(iris.Map{"message": "User deleted successfully"})
}
