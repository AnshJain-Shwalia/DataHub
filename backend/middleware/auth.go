package middleware

import (
	"net/http"
	"strings"

	"github.com/AnshJain-Shwalia/DataHub/backend/config"
	"github.com/AnshJain-Shwalia/DataHub/backend/http_util"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the structure of our JWT claims
type JWTClaims struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// RequireJWT is a middleware that validates JWT tokens and extracts user information
// It adds the user ID to the gin context for downstream handlers to use
func RequireJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, http_util.NewErrorResponse(http.StatusUnauthorized, "Authorization header required", nil))
			c.Abort()
			return
		}

		// Check if it starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, http_util.NewErrorResponse(http.StatusUnauthorized, "Authorization header must start with 'Bearer '", nil))
			c.Abort()
			return
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, http_util.NewErrorResponse(http.StatusUnauthorized, "JWT token required", nil))
			c.Abort()
			return
		}

		// Parse and validate the token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(config.LoadConfig().JWTSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, http_util.NewErrorResponse(http.StatusUnauthorized, "Invalid JWT token", err.Error()))
			c.Abort()
			return
		}

		// Extract claims
		if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
			// Add user information to context for downstream handlers
			c.Set("userID", claims.ID)
			c.Set("userEmail", claims.Email)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, http_util.NewErrorResponse(http.StatusUnauthorized, "Invalid JWT claims", nil))
			c.Abort()
			return
		}
	}
}

// GetUserIDFromContext is a helper function to extract user ID from gin context
// Returns empty string if user ID is not found or invalid
func GetUserIDFromContext(c *gin.Context) string {
	if userID, exists := c.Get("userID"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// GetUserEmailFromContext is a helper function to extract user email from gin context
// Returns empty string if user email is not found
func GetUserEmailFromContext(c *gin.Context) string {
	if userEmail, exists := c.Get("userEmail"); exists {
		if email, ok := userEmail.(string); ok {
			return email
		}
	}
	return ""
}