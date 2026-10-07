package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"Server/database"
	"Server/models"
	"Server/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func setupTestApp() *fiber.App {
	_ = godotenv.Load()
	if os.Getenv("MONGO_URI") == "" {
		os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	}
	os.Setenv("DB_NAME", "social_api_test")
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "test_jwt_secret_123456")
	}

	database.Connect()

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

func doRequest(app *fiber.App, method, url string, body interface{}, token string) (*http.Response, []byte, error) {
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

func TestAllAPIs(t *testing.T) {
	app := setupTestApp()
	if database.DB == nil {
		t.Fatal("Database connection failed, DB is nil")
	}
	// Clean test database before run
	_ = database.DB.Drop(nil)

	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())
	user1Email := fmt.Sprintf("alice_%s@example.com", uniqueSuffix)
	user2Email := fmt.Sprintf("bob_%s@example.com", uniqueSuffix)

	var token1, token2 string
	var user1ID, user2ID string
	var postID string

	// 1. GET /
	t.Run("1. GET / (Welcome)", func(t *testing.T) {
		resp, body, err := doRequest(app, "GET", "/", nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d, body: %s", resp.StatusCode, string(body))
		}
		if string(body) != "Welcome bro!" {
			t.Fatalf("Unexpected body: %s", string(body))
		}
		t.Log("Root endpoint works!")
	})

	// 2. POST /user/signup - Register User 1
	t.Run("2. POST /user/signup (Register User 1)", func(t *testing.T) {
		payload := map[string]string{
			"email":     user1Email,
			"password":  "secret123",
			"firstName": "Alice",
			"lastName":  "Wonderland",
		}
		resp, body, err := doRequest(app, "POST", "/user/signup", payload, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected 200/201, got %d: %s", resp.StatusCode, string(body))
		}

		var res struct {
			Result models.UserModel `json:"result"`
			Token  string           `json:"token"`
		}
		if err := json.Unmarshal(body, &res); err != nil {
			t.Fatalf("Unmarshal error: %v, body: %s", err, string(body))
		}
		token1 = res.Token
		user1ID = res.Result.ID.Hex()
		if user1ID == "" || token1 == "" {
			t.Fatalf("Expected valid user ID and token, got ID: %s, token: %s", user1ID, token1)
		}
		t.Logf("Registered User 1 successfully: ID=%s", user1ID)
	})

	// 3. POST /user/signup - Register User 2
	t.Run("3. POST /user/signup (Register User 2)", func(t *testing.T) {
		payload := map[string]string{
			"email":     user2Email,
			"password":  "secret456",
			"firstName": "Bob",
			"lastName":  "Builder",
		}
		resp, body, err := doRequest(app, "POST", "/user/signup", payload, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected 200/201, got %d: %s", resp.StatusCode, string(body))
		}

		var res struct {
			Result models.UserModel `json:"result"`
			Token  string           `json:"token"`
		}
		if err := json.Unmarshal(body, &res); err != nil {
			t.Fatalf("Unmarshal error: %v, body: %s", err, string(body))
		}
		token2 = res.Token
		user2ID = res.Result.ID.Hex()
		t.Logf("Registered User 2 successfully: ID=%s", user2ID)
	})

	// 4. POST /user/signin - Login User 1
	t.Run("4. POST /user/signin (Login User 1)", func(t *testing.T) {
		payload := map[string]string{
			"email":    user1Email,
			"password": "secret123",
		}
		resp, body, err := doRequest(app, "POST", "/user/signin", payload, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, string(body))
		}

		var res struct {
			Result models.UserModel `json:"result"`
			Token  string           `json:"token"`
		}
		if err := json.Unmarshal(body, &res); err != nil {
			t.Fatalf("Unmarshal error: %v, body: %s", err, string(body))
		}
		if res.Token == "" {
			t.Fatal("Login did not return token")
		}
		t.Log("User 1 login successful")
	})

	// 5. GET /user/getUser/:id - Get User 1 profile
	t.Run("5. GET /user/getUser/:id", func(t *testing.T) {
		resp, body, err := doRequest(app, "GET", "/user/getUser/"+user1ID, nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, string(body))
		}
		var res struct {
			User  models.UserModel   `json:"user"`
			Posts []models.PostModel `json:"posts"`
		}
		if err := json.Unmarshal(body, &res); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if res.User.Email != user1Email {
			t.Fatalf("Expected email %s, got %s", user1Email, res.User.Email)
		}
		t.Log("Fetched user profile successfully")
	})

	// 6. PATCH /user/update/:id - Update User 1 profile
	t.Run("6. PATCH /user/update/:id", func(t *testing.T) {
		payload := map[string]string{
			"name":     "Alice Updated",
			"imageUrl": "https://example.com/avatar.png",
			"bio":      "Software Engineer at Synoebook",
		}
		resp, body, err := doRequest(app, "PATCH", "/user/update/"+user1ID, payload, token1)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, string(body))
		}
		t.Log("Updated user profile successfully")
	})

	// 7. PATCH /user/:id/following - User 1 follows User 2
	t.Run("7. PATCH /user/:id/following (Follow User 2)", func(t *testing.T) {
		resp, body, err := doRequest(app, "PATCH", "/user/"+user2ID+"/following", nil, token1)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, string(body))
		}
		t.Log("User 1 followed User 2 successfully")
	})

	// 8. GET /user/getSug - Get suggestions for User 1
	t.Run("8. GET /user/getSug", func(t *testing.T) {
		resp, body, err := doRequest(app, "GET", "/user/getSug?id="+user1ID, nil, token1)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, string(body))
		}
		t.Log("Get suggestions succeeded")
	})

	// 9. POST /posts - User 1 creates post
	t.Run("9. POST /posts (Create Post)", func(t *testing.T) {
		payload := map[string]string{
			"title":        "First Synoebook Post",
			"message":      "Hello world, testing the new Golang Fiber backend!",
			"selectedFile": "https://example.com/cover.jpg",
		}
		resp, body, err := doRequest(app, "POST", "/posts", payload, token1)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 201, got %d: %s", resp.StatusCode, string(body))
		}

		var createdPost models.PostModel
		if err := json.Unmarshal(body, &createdPost); err != nil {
			t.Fatalf("Unmarshal post error: %v, body: %s", err, string(body))
		}
		postID = createdPost.ID.Hex()
		if postID == "" {
			t.Fatalf("Created post ID is empty")
		}
		t.Logf("Created post successfully: ID=%s", postID)
	})

	// 10. GET /posts - Get all posts
	t.Run("10. GET /posts (Feed posts)", func(t *testing.T) {
		resp, body, err := doRequest(app, "GET", "/posts?id="+user1ID+"&page=1", nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, string(body))
		}
		t.Log("Fetched posts list successfully")
	})

	// 11. GET /posts/:id - Get single post
	t.Run("11. GET /posts/:id", func(t *testing.T) {
		resp, body, err := doRequest(app, "GET", "/posts/"+postID, nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, string(body))
		}
		t.Log("Fetched single post successfully")
	})

	// 12. PATCH /posts/:id - Update post
	t.Run("12. PATCH /posts/:id (Update Post)", func(t *testing.T) {
		payload := map[string]string{
			"title":        "Updated Post Title",
			"message":      "Updated message content here!",
			"selectedFile": "https://example.com/updated.jpg",
		}
		resp, body, err := doRequest(app, "PATCH", "/posts/"+postID, payload, token1)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected 200/201, got %d: %s", resp.StatusCode, string(body))
		}
		t.Log("Updated post successfully")
	})

	// 13. POST /posts/:id/commentPost - User 2 comments on post
	t.Run("13. POST /posts/:id/commentPost", func(t *testing.T) {
		payload := map[string]string{
			"value": "Great post Alice!",
		}
		resp, body, err := doRequest(app, "POST", "/posts/"+postID+"/commentPost", payload, token2)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, string(body))
		}
		t.Log("Commented on post successfully")
	})

	// 14. PATCH /posts/:id/likePost - User 2 likes post
	t.Run("14. PATCH /posts/:id/likePost", func(t *testing.T) {
		resp, body, err := doRequest(app, "PATCH", "/posts/"+postID+"/likePost", nil, token2)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, string(body))
		}
		t.Log("Liked post successfully")
	})

	// 15. GET /posts/search - Search posts and users
	t.Run("15. GET /posts/search", func(t *testing.T) {
		resp, body, err := doRequest(app, "GET", "/posts/search?searchQuery=Alice", nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, string(body))
		}
		t.Log("Searched posts and users successfully")
	})

	// 16. POST /chat/sendmessage - User 1 sends message to User 2
	t.Run("16. POST /chat/sendmessage", func(t *testing.T) {
		payload := map[string]string{
			"sender":  user1ID,
			"recever": user2ID,
			"content": "Hey Bob! How are you doing?",
		}
		resp, body, err := doRequest(app, "POST", "/chat/sendmessage", payload, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected 201, got %d: %s", resp.StatusCode, string(body))
		}
		t.Log("Chat message sent successfully")
	})

	// 17. GET /chat/getmsgsbynums - Get chat history
	t.Run("17. GET /chat/getmsgsbynums", func(t *testing.T) {
		url := fmt.Sprintf("/chat/getmsgsbynums?from=0&firstuid=%s&seconduid=%s", user1ID, user2ID)
		resp, body, err := doRequest(app, "GET", url, nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, string(body))
		}
		t.Log("Fetched chat messages successfully")
	})

	// 18. GET /chat/get-user-unreadedmsg - Get unread count for User 2
	t.Run("18. GET /chat/get-user-unreadedmsg", func(t *testing.T) {
		resp, body, err := doRequest(app, "GET", "/chat/get-user-unreadedmsg?userid="+user2ID, nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, string(body))
		}
		t.Log("Fetched unread message count successfully")
	})

	// 19. GET /chat/mark-msg-asreaded - Mark messages as read
	t.Run("19. GET /chat/mark-msg-asreaded", func(t *testing.T) {
		url := fmt.Sprintf("/chat/mark-msg-asreaded?mainuid=%s&otheruid=%s", user2ID, user1ID)
		resp, body, err := doRequest(app, "GET", url, nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, string(body))
		}
		t.Log("Marked message as read successfully")
	})

	// 20. GET /notification/:userid - Get User 2 notifications
	t.Run("20. GET /notification/:userid", func(t *testing.T) {
		resp, body, err := doRequest(app, "GET", "/notification/"+user2ID, nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, string(body))
		}
		t.Log("Retrieved notifications successfully")
	})

	// 21. GET /notification/mark-notification-asreaded
	t.Run("21. GET /notification/mark-notification-asreaded", func(t *testing.T) {
		resp, body, err := doRequest(app, "GET", "/notification/mark-notification-asreaded?id="+user2ID, nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, string(body))
		}
		t.Log("Marked notifications as read successfully")
	})

	// 22. DELETE /posts/:id - User 1 deletes their post
	t.Run("22. DELETE /posts/:id (Delete Post)", func(t *testing.T) {
		resp, body, err := doRequest(app, "DELETE", "/posts/"+postID, nil, token1)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, string(body))
		}
		t.Log("Post deleted successfully")
	})

	// 23. DELETE /user/delete/:id - User 1 deletes account
	t.Run("23. DELETE /user/delete/:id (Delete User)", func(t *testing.T) {
		resp, body, err := doRequest(app, "DELETE", "/user/delete/"+user1ID, nil, token1)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, string(body))
		}
		t.Log("User deleted successfully")
	})
}
