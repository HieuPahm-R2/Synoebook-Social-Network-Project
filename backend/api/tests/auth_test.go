package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"Server/models"
)

func TestAuthModule(t *testing.T) {
	app := testApp
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("auth_user_%s@example.com", uniqueSuffix)
	password := "securePass123"

	t.Run("SignUp - Successful Registration", func(t *testing.T) {
		payload := models.CreateUser{
			Email:     email,
			Password:  password,
			FirstName: "Auth",
			LastName:  "Tester",
		}

		resp, body, err := MakeRequest(app, "POST", "/user/signup", payload, "")
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected 200 or 201, got %d, body: %s", resp.StatusCode, string(body))
		}

		var res struct {
			Result models.UserModel `json:"result"`
			Token  string           `json:"token"`
		}
		if err := json.Unmarshal(body, &res); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if res.Result.Email != email {
			t.Errorf("Expected email %s, got %s", email, res.Result.Email)
		}
		if res.Result.Name != "Auth Tester" {
			t.Errorf("Expected name 'Auth Tester', got %s", res.Result.Name)
		}
		if res.Token == "" {
			t.Error("Expected non-empty JWT token")
		}
	})

	t.Run("SignUp - Duplicate Email Failure", func(t *testing.T) {
		payload := models.CreateUser{
			Email:     email,
			Password:  password,
			FirstName: "Duplicate",
			LastName:  "User",
		}

		resp, _, err := MakeRequest(app, "POST", "/user/signup", payload, "")
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request for duplicate email, got %d", resp.StatusCode)
		}
	})

	t.Run("SignUp - Validation Failure (Password too short)", func(t *testing.T) {
		payload := models.CreateUser{
			Email:     fmt.Sprintf("short_%s@example.com", uniqueSuffix),
			Password:  "123", // min is 5
			FirstName: "Short",
			LastName:  "Pass",
		}

		resp, _, err := MakeRequest(app, "POST", "/user/signup", payload, "")
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request for short password, got %d", resp.StatusCode)
		}
	})

	t.Run("SignIn - Successful Login", func(t *testing.T) {
		payload := models.LoginUser{
			Email:    email,
			Password: password,
		}

		resp, body, err := MakeRequest(app, "POST", "/user/signin", payload, "")
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d, body: %s", resp.StatusCode, string(body))
		}

		var res struct {
			Result models.UserModel `json:"result"`
			Token  string           `json:"token"`
		}
		if err := json.Unmarshal(body, &res); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if res.Result.Email != email {
			t.Errorf("Expected email %s, got %s", email, res.Result.Email)
		}
		if res.Token == "" {
			t.Error("Expected non-empty JWT token on login")
		}
	})

	t.Run("SignIn - Wrong Password", func(t *testing.T) {
		payload := models.LoginUser{
			Email:    email,
			Password: "wrongPassword999",
		}

		resp, _, err := MakeRequest(app, "POST", "/user/signin", payload, "")
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request for wrong password, got %d", resp.StatusCode)
		}
	})

	t.Run("SignIn - Non-existent User", func(t *testing.T) {
		payload := models.LoginUser{
			Email:    "non_existent_email_123456@example.com",
			Password: "anyPassword123",
		}

		resp, _, err := MakeRequest(app, "POST", "/user/signin", payload, "")
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request for non-existent user, got %d", resp.StatusCode)
		}
	})
}
