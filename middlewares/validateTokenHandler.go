package middlewares

import (
	"net/http"
	"strings"

	"node_management_application/config"

	"github.com/dgrijalva/jwt-go"
	"github.com/kataras/iris/v12"
)

// UserClaims defines the structure of the JWT payload
type UserClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.StandardClaims
}

// Authenticate is the JWT middleware to validate tokens
func Authenticate(ctx iris.Context) {
	// Extract the token from the Authorization header
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Authorization token required"})
		return
	}

	// Check if the Authorization header has the "Bearer " prefix
	if !strings.HasPrefix(authHeader, "Bearer ") {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Invalid Authorization format"})
		return
	}

	// Remove the "Bearer " prefix
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Parse and validate the token
	claims := &UserClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return config.JWTSecretKey, nil
	})

	if err != nil || !token.Valid {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "Invalid or expired token"})
		return
	}

	// Store user information in the context for downstream handlers
	ctx.Values().Set("user_id", claims.UserID)
	ctx.Values().Set("email", claims.Email)

	// Proceed to the next handler
	ctx.Next()
}
