package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/AnshJain-Shwalia/DataHub/backend/config"
	"github.com/gin-gonic/gin/binding"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// createGoogleAuth creates and returns a new Google OAuth config with the specified redirect URL
func createGoogleAuth(redirectURL string) *oauth2.Config {
	envCfg := config.LoadConfig()
	return &oauth2.Config{
		ClientID:     envCfg.GoogleClientID,
		ClientSecret: envCfg.GoogleClientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

// GoogleUserInfo represents the essential user information returned by Google's userinfo endpoint
type GoogleUserInfo struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name" binding:"required"`
}

// GetUserInfo retrieves user info from Google's userinfo API using the provided access token
func GetUserInfo(token *oauth2.Token) (*GoogleUserInfo, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a client with the token
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))

	// Make request to Google's userinfo endpoint
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}
	defer resp.Body.Close()

	// Check if the response status code is OK
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("userinfo request failed with status %d: %s", resp.StatusCode, body)
	}

	// Parse the JSON response
	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %v", err)
	}

	// Validate using Gin's validator
	if err := binding.Validator.ValidateStruct(userInfo); err != nil {
		return nil, fmt.Errorf("invalid user info: %v", err)
	}

	return &userInfo, nil
}

// GetUserInfoFromAccessToken retrieves user info using just the access token string
func GetUserInfoFromAccessToken(accessToken string) (*GoogleUserInfo, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a request
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add the authorization header
	req.Header.Add("Authorization", "Bearer "+accessToken)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}
	defer resp.Body.Close()
	resp.Cookies()
	// Check if the response status code is OK
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("userinfo request failed with status %d: %s", resp.StatusCode, body)
	}

	// Parse the JSON response
	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %v", err)
	}

	// Validate using Gin's validator
	if err := binding.Validator.ValidateStruct(userInfo); err != nil {
		return nil, fmt.Errorf("invalid user info: %v", err)
	}

	return &userInfo, nil
}

// ExchangeCodeForTokens exchanges auth code for tokens using the Google library
func ExchangeCodeForTokens(code string) (*oauth2.Token, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a new config with default redirect URL
	oauthConfig := createGoogleAuth("http://localhost:3000/auth/google/callback")

	// Exchange will handle all the HTTP details and parameter encoding
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %v", err)
	}

	// Validate the token
	if !token.Valid() {
		return nil, fmt.Errorf("received invalid token from Google")
	}

	fmt.Println("====TOKEN====")
	fmt.Println(token)
	fmt.Println("====TOKEN====")

	return token, nil
}

// GenerateOAuthURL generates the OAuth2 URL for Google login
func GenerateOAuthURL(state string) string {
	oauthConfig := createGoogleAuth("http://localhost:3000/auth/google/callback")
	return oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// GenerateOAuthUrlWithRedirectUrl constructs a Google OAuth2 authorization URL using the provided state and redirectURL.
// If redirectURL is empty, it defaults to "http://localhost:3000/auth/google/callback".
// This URL can be sent to the frontend to initiate the OAuth login flow.
func GenerateOAuthUrlWithRedirectUrl(state string, redirectURL string) string {
	if redirectURL == "" {
		redirectURL = "http://localhost:3000/auth/google/callback"
	}
	oauthConfig := createGoogleAuth(redirectURL)
	// Pass state for CSRF protection, request offline access and force approval prompt
	return oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}
