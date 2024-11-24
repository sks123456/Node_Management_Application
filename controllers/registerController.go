package controllers

import (
	"net/http"

	"node_management_application/config"
	"node_management_application/models"

	"github.com/kataras/iris/v12"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(ctx iris.Context) {
	// Parse user registration details
	var user struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := ctx.ReadJSON(&user); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request body"})
		return
	}

	// Check if email is already registered
	var existingUser models.User
	if result := config.DB.Where("email = ?", user.Email).First(&existingUser); result.RowsAffected > 0 {
		ctx.StatusCode(http.StatusConflict)
		ctx.JSON(iris.Map{"error": "Email is already registered"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to hash password"})
		return
	}

	// Save the new user
	newUser := models.User{
		Name:     user.Name,
		Email:    user.Email,
		Password: string(hashedPassword),
	}
	if result := config.DB.Create(&newUser); result.Error != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create user"})
		return
	}

	// Respond with the user detail created
	ctx.JSON(iris.Map{
		"user":  newUser,
	})
}
