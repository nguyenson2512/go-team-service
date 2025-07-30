package response

import (
	"github.com/gin-gonic/gin"
)

// Success sends a successful response
func Success(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

// Error sends an error response
func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{"error": message})
}

// ErrorWithDetails sends an error response with additional details
func ErrorWithDetails(c *gin.Context, statusCode int, message string, details interface{}) {
	c.JSON(statusCode, gin.H{
		"error":   message,
		"details": details,
	})
}
