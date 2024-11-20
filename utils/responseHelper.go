package utils

import (
	"net/http"

	"github.com/kataras/iris/v12"
)

// ValidationErrorResponse sends a standardized validation error response
func ValidationErrorResponse(ctx iris.Context, err error) {
	ctx.StatusCode(http.StatusBadRequest)
	ctx.JSON(iris.Map{
		"error":  "Validation failed",
		"detail": err.Error(),
	})
}
