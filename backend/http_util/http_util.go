package http_util

import "github.com/gin-gonic/gin"

func NewErrorResponse(statusCode int, message string, details interface{}) gin.H {
	if message == "" {
		switch statusCode {
		case 400:
			message = "Bad Request"
		case 401:
			message = "Unauthorized"
		case 403:
			message = "Forbidden"
		case 404:
			message = "Not Found"
		case 500:
			message = "Internal Server Error"
		default:
			message = "An error occurred"
		}
	}

	response := gin.H{
		"error": gin.H{
			"message": message,
		},
		"success": false,
		"status":  statusCode,
	}

	if details != nil {
		response["error"].(gin.H)["details"] = details
	}

	return response
}
