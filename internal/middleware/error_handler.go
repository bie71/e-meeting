package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			log.Error().Err(err).Msg("Request error")

			// Determine appropriate status code
			statusCode := http.StatusInternalServerError
			switch err.(type) {
			case *gin.Error:
				statusCode = http.StatusBadRequest
			}

			c.JSON(statusCode, ErrorResponse{
				Error: err.Error(),
			})
		}
	}
}
