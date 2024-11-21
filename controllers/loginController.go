package controllers

import (
	"net/http"
	"time"

	"node_management_application/config"
	"node_management_application/models"

	"golang.org/x/crypto/bcrypt"

	"github.com/dgrijalva/jwt-go"
	"github.com/kataras/iris/v12"
)

// UserClaims defines the structure of the JWT payload
type UserClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.StandardClaims
}

// Login handles user authentication and token generation
func Login(ctx iris.Context) {
	// Read login credentials from the request body
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := ctx.ReadJSON(&credentials); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request"})
		return
	}

	// Fetch the user from the database
	var user models.User
	result := config.DB.Where("email = ?", credentials.Email).First(&user)
	if result.Error != nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Invalid email or password"})
		return
	}

	// Compare the hashed password with the provided password
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
	if err != nil {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Invalid email or password"})
		return
	}

	// Create the JWT token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &UserClaims{
		UserID: user.ID,
		Email:  user.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(config.JWTSecretKey)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to generate token"})
		return
	}

	// Return the token and user information
	ctx.JSON(iris.Map{
		"user": iris.Map{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
		"token": tokenString,
	})
}
