package clients

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

type UserResponse struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	RoleID    int    `json:"role_id"`
	Org       *Org   `json:"org,omitempty"`
}

type Org struct {
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	CompanyDescription string `json:"description"`
	Size               string `json:"size"`
	Industry           string `json:"industry"`
	Domain             string `json:"domain"`
}

type UserLoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// ValidateToken validates a token and checks if the user is authorized for a given action
func ValidateToken(token string, action string) (*UserResponse, error) {
	// Get auth service URL from environment variable
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		return nil, errors.New("AUTH_SERVICE_URL environment variable is not set")
	}

	// First validate the token
	payload := map[string]string{
		"token":  token,
		"action": action,
	}
	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", authServiceURL+"/auth/validate", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create validation request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth validation failed with status: %s", resp.Status)
	}

	// Then get user details
	req, err = http.NewRequest("GET", authServiceURL+"/auth/get_user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user details: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user details with status: %s", resp.Status)
	}

	var userResp UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return nil, fmt.Errorf("failed to decode user response: %v", err)
	}

	// Check if user has an organization
	if userResp.Org == nil {
		return nil, errors.New("user does not have an associated organization")
	}

	return &userResp, nil
}

// GetCompanyID returns the organization ID from the user response
func GetCompanyID(user *UserResponse) (int, error) {
	if user == nil || user.Org == nil {
		return 0, errors.New("user or organization is nil")
	}
	return user.Org.ID, nil
}
