package auth

import (
	"fmt"
	"net/http"

	"github.com/AnshJain-Shwalia/DataHub/backend/http_util"
	"github.com/gin-gonic/gin"
)

type AuthCodeRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}

type GenerateOAuthURLRequest struct {
	RedirectURL string `json:"redirectURL"`
}

func AuthCodeHandler(c *gin.Context) {
	var body AuthCodeRequest
	if err := c.ShouldBindBodyWithJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Incorrect body structure."})
		return
	}

	// Check state BEFORE processing the code
	if !VerifyAndConsumeState(body.State) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	token, err := ExchangeCodeForTokens(body.Code)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Problem in exchanging code for token."})
	}
	fmt.Print(token)
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func GenerateOAuthURLHandler(c *gin.Context) {
	state, err := GenerateAndAddState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, http_util.NewErrorResponse(http.StatusInternalServerError, "", nil))
		return
	}
	var body GenerateOAuthURLRequest
	if err := c.ShouldBindBodyWithJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	authURL := GenerateOAuthUrlWithRedirectUrl(state, body.RedirectURL)
	c.JSON(http.StatusOK, gin.H{
		"authURL": authURL,
	})
}
