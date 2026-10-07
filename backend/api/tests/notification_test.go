package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"Server/models"
)

func TestNotificationModule(t *testing.T) {
	app := testApp
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	_, token1 := RegisterAndLogin(t, app,
		fmt.Sprintf("notif_user_1_%s@example.com", uniqueSuffix),
		"password123",
		"NotifSender", "One")

	user2ID, _ := RegisterAndLogin(t, app,
		fmt.Sprintf("notif_user_2_%s@example.com", uniqueSuffix),
		"password456",
		"NotifReceiver", "Two")

	// Trigger a notification for User 2 by having User 1 follow User 2
	_, _, err := MakeRequest(app, "PATCH", "/user/"+user2ID+"/following", nil, token1)
	if err != nil {
		t.Fatalf("Failed to trigger follow notification: %v", err)
	}

	t.Run("GetUserNotification - Retrieve Notifications", func(t *testing.T) {
		resp, body, err := MakeRequest(app, "GET", "/notification/"+user2ID, nil, "")
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d, body: %s", resp.StatusCode, string(body))
		}

		var res struct {
			Notifications []models.Notification `json:"notifications"`
		}
		if err := json.Unmarshal(body, &res); err != nil {
			t.Fatalf("Failed to parse notifications response: %v", err)
		}

		if len(res.Notifications) == 0 {
			t.Error("Expected at least 1 notification for User 2")
		}
	})

	t.Run("MarknotAsReaded - Mark Notifications As Read", func(t *testing.T) {
		resp, body, err := MakeRequest(app, "GET", "/notification/mark-notification-asreaded?id="+user2ID, nil, "")
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d, body: %s", resp.StatusCode, string(body))
		}

		var res struct {
			Notifications []models.Notification `json:"notifications"`
		}
		if err := json.Unmarshal(body, &res); err != nil {
			t.Fatalf("Failed to parse mark response: %v", err)
		}

		for _, notif := range res.Notifications {
			if !notif.IsReaded {
				t.Errorf("Expected notification %s to be marked as read", notif.ID.Hex())
			}
		}
	})
}
