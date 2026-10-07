package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"Server/database"
	"Server/models"
	"Server/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

var testApp *fiber.App

func TestMain(m *testing.M) {
	_ = godotenv.Load("../.env")
	if os.Getenv("MONGO_URI") == "" {
		_ = os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	}
	_ = os.Setenv("DB_NAME", "social_module_test")
	if os.Getenv("JWT_SECRET") == "" {
		_ = os.Setenv("JWT_SECRET", "super_secret_test_jwt_key_2026")
	}

	database.Connect()
	if database.DB == nil {
		fmt.Println("CRITICAL: Failed to connect to MongoDB in TestMain")
		os.Exit(1)
	}

	// Clean up database before tests
	_ = database.DB.Drop(context.Background())

	testApp = SetupTestApp()

	code := m.Run()

	// Clean up database after tests
	_ = database.DB.Drop(context.Background())

	os.Exit(code)
}

func SetupTestApp() *fiber.App {
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOriginsFunc: func(origin string) bool { return true },
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome bro!")
	})

	routes.SetupAuthRoutes(app)
	routes.SetupUserRoutes(app)
	routes.SetupPostRoutes(app)
	routes.SetupChatRoutes(app)
	routes.SetupNotificationRoutes(app)

	return app
}

func MakeRequest(app *fiber.App, method, url string, body interface{}, token string) (*http.Response, []byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		bodyReader = bytes.NewReader(jsonBytes)
	}

	req := httptest.NewRequest(method, url, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := app.Test(req, 10000)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	return resp, respBody, err
}

func RegisterAndLogin(t *testing.T, app *fiber.App, email, password, firstName, lastName string) (string, string) {
	t.Helper()

	signupPayload := models.CreateUser{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
	}

	resp, body, err := MakeRequest(app, "POST", "/user/signup", signupPayload, "")
	if err != nil {
		t.Fatalf("Failed to execute signup request: %v", err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Fatalf("Signup failed with status %d: %s", resp.StatusCode, string(body))
	}

	var signupRes struct {
		Result models.UserModel `json:"result"`
		Token  string           `json:"token"`
	}
	if err := json.Unmarshal(body, &signupRes); err != nil {
		t.Fatalf("Failed to unmarshal signup response: %v, raw: %s", err, string(body))
	}

	return signupRes.Result.ID.Hex(), signupRes.Token
}
