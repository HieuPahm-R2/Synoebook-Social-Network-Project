package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"Server/models"
)

func TestUserModule(t *testing.T) {
	app := testApp
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	user1ID, token1 := RegisterAndLogin(t, app,
		fmt.Sprintf("user_mod_1_%s@example.com", uniqueSuffix),
		"password123",
		"UserOne", "Test")

	user2ID, token2 := RegisterAndLogin(t, app,
		fmt.Sprintf("user_mod_2_%s@example.com", uniqueSuffix),
		"password456",
		"UserTwo", "Test")

	t.Run("GetUserByID - Success", func(t *testing.T) {
		resp, body, err := MakeRequest(app, "GET", "/user/getUser/"+user1ID, nil, "")
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d, body: %s", resp.StatusCode, string(body))
		}

		var res struct {
			User  models.UserModel   `json:"user"`
			Posts []models.PostModel `json:"posts"`
		}
		if err := json.Unmarshal(body, &res); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if res.User.Name != "UserOne Test" {
			t.Errorf("Expected user name 'UserOne Test', got '%s'", res.User.Name)
		}
	})

	t.Run("GetUserByID - Invalid ID", func(t *testing.T) {
		resp, _, err := MakeRequest(app, "GET", "/user/getUser/000000000000000000000000", nil, "")
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode == http.StatusOK {
			t.Error("Expected error status for non-existent user ID, got 200 OK")
		}
	})

	t.Run("UpdateUser - Success with own token", func(t *testing.T) {
		payload := models.UpdateUser{
			Name:     "UserOne Updated",
			ImageUrl: "https://example.com/photo.png",
			Bio:      "Bio of User One",
		}

		resp, body, err := MakeRequest(app, "PATCH", "/user/update/"+user1ID, payload, token1)
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d, body: %s", resp.StatusCode, string(body))
		}

		var res struct {
			Data models.UserModel `json:"data"`
		}
		if err := json.Unmarshal(body, &res); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if res.Data.Name != "UserOne Updated" {
			t.Errorf("Expected updated name 'UserOne Updated', got '%s'", res.Data.Name)
		}
		if res.Data.Bio != "Bio of User One" {
			t.Errorf("Expected updated bio 'Bio of User One', got '%s'", res.Data.Bio)
		}
	})

	t.Run("UpdateUser - Unauthorized (Updating another user)", func(t *testing.T) {
		payload := models.UpdateUser{
			Name: "Hacked Name",
		}

		resp, _, err := MakeRequest(app, "PATCH", "/user/update/"+user1ID, payload, token2)
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized, got %d", resp.StatusCode)
		}
	})

	t.Run("FollowingUser - Follow then Unfollow toggle", func(t *testing.T) {
		// Follow User 2
		respFollow, bodyFollow, err := MakeRequest(app, "PATCH", "/user/"+user2ID+"/following", nil, token1)
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if respFollow.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK on follow, got %d, body: %s", respFollow.StatusCode, string(bodyFollow))
		}

		var followRes struct {
			FirstUser  models.UserModel `json:"FirstUser"`
			SecondUser models.UserModel `json:"SecondUser"`
		}
		if err := json.Unmarshal(bodyFollow, &followRes); err != nil {
			t.Fatalf("Failed to parse follow response: %v", err)
		}

		// User 2 (FirstUser in response) should now have User 1 in followers
		foundFollower := false
		for _, follower := range followRes.FirstUser.Followers {
			if follower == user1ID {
				foundFollower = true
				break
			}
		}
		if !foundFollower {
			t.Error("User 1 ID should be present in User 2 followers list")
		}

		// Unfollow User 2 (calling same endpoint toggles unfollow)
		respUnfollow, bodyUnfollow, err := MakeRequest(app, "PATCH", "/user/"+user2ID+"/following", nil, token1)
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if respUnfollow.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK on unfollow, got %d, body: %s", respUnfollow.StatusCode, string(bodyUnfollow))
		}

		var unfollowRes struct {
			FirstUser models.UserModel `json:"FirstUser"`
		}
		if err := json.Unmarshal(bodyUnfollow, &unfollowRes); err != nil {
			t.Fatalf("Failed to parse unfollow response: %v", err)
		}
		for _, follower := range unfollowRes.FirstUser.Followers {
			if follower == user1ID {
				t.Error("User 1 ID should NOT be present in User 2 followers after unfollowing")
			}
		}
	})

	t.Run("GetSugUser - Suggested Users", func(t *testing.T) {
		resp, body, err := MakeRequest(app, "GET", "/user/getSug?id="+user1ID, nil, token1)
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d, body: %s", resp.StatusCode, string(body))
		}

		var res struct {
			Users []models.UserModel `json:"users"`
		}
		if err := json.Unmarshal(body, &res); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
		if res.Users == nil {
			t.Error("Expected non-nil slice of suggested users")
		}
	})

	t.Run("DeleteUser - Unauthorized Attempt", func(t *testing.T) {
		resp, _, err := MakeRequest(app, "DELETE", "/user/delete/"+user1ID, nil, token2)
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized when deleting other user, got %d", resp.StatusCode)
		}
	})

	t.Run("DeleteUser - Success with own token", func(t *testing.T) {
		resp, body, err := MakeRequest(app, "DELETE", "/user/delete/"+user1ID, nil, token1)
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK on delete user, got %d, body: %s", resp.StatusCode, string(body))
		}
	})
}
