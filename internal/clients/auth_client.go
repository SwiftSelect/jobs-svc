package clients

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
	CompanyDescription string `json:"companyDescription"`
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
	payload := map[string]string{
		"token":  token,
		"action": action,
	}
	payloadBytes, _ := json.Marshal(payload)

	resp, err := http.Post("http://localhost:8005/auth/validate", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, errors.New("failed to connect to auth microservice")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth validation failed: %s", resp.Status)
	}

	var userResp UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return nil, errors.New("failed to decode auth response")
	}

	return &userResp, nil
}
