package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"Server/models"
)

func TestPostModule(t *testing.T) {
	app := testApp
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	user1ID, token1 := RegisterAndLogin(t, app,
		fmt.Sprintf("post_user_1_%s@example.com", uniqueSuffix),
		"password123",
		"PostAuthor", "One")

	_, token2 := RegisterAndLogin(t, app,
		fmt.Sprintf("post_user_2_%s@example.com", uniqueSuffix),
		"password456",
		"PostCommenter", "Two")

	var postID string

	t.Run("CreatePost - Success", func(t *testing.T) {
		payload := models.CreateOrUpdatePost{
			Title:        "Golang Fiber & MongoDB Architecture",
			Message:      "Exploring high performance social networking backend with Go Fiber.",
			SelectedFile: "https://example.com/golang.jpg",
		}

		resp, body, err := MakeRequest(app, "POST", "/posts", payload, token1)
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 201 or 200, got %d, body: %s", resp.StatusCode, string(body))
		}

		var createdPost models.PostModel
		if err := json.Unmarshal(body, &createdPost); err != nil {
			t.Fatalf("Failed to parse post response: %v", err)
		}

		if createdPost.ID.IsZero() {
			t.Fatal("Expected created post to have a non-zero ID")
		}
		if createdPost.Title != payload.Title {
			t.Errorf("Expected title '%s', got '%s'", payload.Title, createdPost.Title)
		}
		if createdPost.Creator != user1ID {
			t.Errorf("Expected creator to be %s, got %s", user1ID, createdPost.Creator)
		}

		postID = createdPost.ID.Hex()
	})

	t.Run("CreatePost - Unauthorized (Missing Token)", func(t *testing.T) {
		payload := models.CreateOrUpdatePost{
			Title:   "Unauthenticated Post",
			Message: "This should fail because no token was provided.",
		}

		resp, _, err := MakeRequest(app, "POST", "/posts", payload, "")
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized, got %d", resp.StatusCode)
		}
	})

	t.Run("CreatePost - Validation Failure (Short Message)", func(t *testing.T) {
		payload := models.CreateOrUpdatePost{
			Title:   "Invalid Post",
			Message: "Hi", // Less than 5 characters
		}

		resp, _, err := MakeRequest(app, "POST", "/posts", payload, token1)
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request for short message, got %d", resp.StatusCode)
		}
	})

	t.Run("GetAllPosts - Feed with Pagination", func(t *testing.T) {
		resp, body, err := MakeRequest(app, "GET", "/posts?id="+user1ID+"&page=1", nil, "")
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d, body: %s", resp.StatusCode, string(body))
		}

		var res struct {
			Data        []models.PostModel `json:"data"`
			CurrentPage int                `json:"currentPage"`
		}
		if err := json.Unmarshal(body, &res); err != nil {
			t.Fatalf("Failed to parse posts response: %v", err)
		}
		if res.CurrentPage != 1 {
			t.Errorf("Expected currentPage 1, got %d", res.CurrentPage)
		}
	})

	t.Run("GetPost - Retrieve By ID", func(t *testing.T) {
		resp, body, err := MakeRequest(app, "GET", "/posts/"+postID, nil, "")
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d, body: %s", resp.StatusCode, string(body))
		}

		var res struct {
			Post models.PostModel `json:"post"`
		}
		if err := json.Unmarshal(body, &res); err != nil {
			t.Fatalf("Failed to parse single post response: %v", err)
		}
		if res.Post.ID.Hex() != postID {
			t.Errorf("Expected post ID %s, got %s", postID, res.Post.ID.Hex())
		}
	})

	t.Run("UpdatePost - Success by Owner", func(t *testing.T) {
		payload := models.CreateOrUpdatePost{
			Title:        "Updated Architecture Title",
			Message:      "Updated post message content goes here.",
			SelectedFile: "https://example.com/updated.jpg",
		}

		resp, body, err := MakeRequest(app, "PATCH", "/posts/"+postID, payload, token1)
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected 200 or 201, got %d, body: %s", resp.StatusCode, string(body))
		}
	})

	t.Run("UpdatePost - Non-owner attempt", func(t *testing.T) {
		payload := models.CreateOrUpdatePost{
			Title:   "Hacked Post",
			Message: "Attempting to overwrite another user's post.",
		}

		resp, _, err := MakeRequest(app, "PATCH", "/posts/"+postID, payload, token2)
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			t.Error("Expected error status when updating another user's post")
		}
	})

	t.Run("CommentPost - Add comment", func(t *testing.T) {
		payload := models.ComnmentPost{
			Value: "Great article on Go Fiber!",
		}

		resp, body, err := MakeRequest(app, "POST", "/posts/"+postID+"/commentPost", payload, token2)
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK on comment, got %d, body: %s", resp.StatusCode, string(body))
		}

		var res struct {
			Data models.PostModel `json:"data"`
		}
		if err := json.Unmarshal(body, &res); err != nil {
			t.Fatalf("Failed to parse comment response: %v", err)
		}
		if len(res.Data.Comments) == 0 {
			t.Error("Expected post comments to have at least 1 comment")
		}
	})

	t.Run("LikePost - Toggle Like and Unlike", func(t *testing.T) {
		// Like post
		respLike, bodyLike, err := MakeRequest(app, "PATCH", "/posts/"+postID+"/likePost", nil, token2)
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if respLike.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK on like, got %d, body: %s", respLike.StatusCode, string(bodyLike))
		}

		var likeRes struct {
			Post models.PostModel `json:"post"`
		}
		if err := json.Unmarshal(bodyLike, &likeRes); err != nil {
			t.Fatalf("Failed to parse like response: %v", err)
		}
		if len(likeRes.Post.Likes) == 0 {
			t.Error("Expected post to have at least 1 like after liking")
		}

		// Unlike post
		respUnlike, bodyUnlike, err := MakeRequest(app, "PATCH", "/posts/"+postID+"/likePost", nil, token2)
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if respUnlike.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK on unlike, got %d, body: %s", respUnlike.StatusCode, string(bodyUnlike))
		}

		var unlikeRes struct {
			Post models.PostModel `json:"post"`
		}
		if err := json.Unmarshal(bodyUnlike, &unlikeRes); err != nil {
			t.Fatalf("Failed to parse unlike response: %v", err)
		}
		if len(unlikeRes.Post.Likes) != 0 {
			t.Errorf("Expected 0 likes after unliking, got %d", len(unlikeRes.Post.Likes))
		}
	})

	t.Run("SearchPosts - Search by Keyword", func(t *testing.T) {
		resp, body, err := MakeRequest(app, "GET", "/posts/search?searchQuery=Architecture", nil, "")
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK on search, got %d, body: %s", resp.StatusCode, string(body))
		}

		var res struct {
			Posts []models.PostModel `json:"posts"`
			User  []models.UserModel `json:"user"`
		}
		if err := json.Unmarshal(body, &res); err != nil {
			t.Fatalf("Failed to parse search response: %v", err)
		}
		if len(res.Posts) == 0 {
			t.Error("Expected to find at least 1 post matching 'Architecture'")
		}
	})

	t.Run("DeletePost - Non-owner attempt", func(t *testing.T) {
		resp, _, err := MakeRequest(app, "DELETE", "/posts/"+postID, nil, token2)
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode == http.StatusOK {
			t.Error("Expected error status when non-owner deletes post, got 200")
		}
	})

	t.Run("DeletePost - Success by Owner", func(t *testing.T) {
		resp, body, err := MakeRequest(app, "DELETE", "/posts/"+postID, nil, token1)
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK on delete post, got %d, body: %s", resp.StatusCode, string(body))
		}
	})
}
