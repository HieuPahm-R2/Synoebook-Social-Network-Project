package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"Server/models"
)

func TestChatModule(t *testing.T) {
	app := testApp
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	user1ID, _ := RegisterAndLogin(t, app,
		fmt.Sprintf("chat_user_1_%s@example.com", uniqueSuffix),
		"password123",
		"ChatSender", "One")

	user2ID, _ := RegisterAndLogin(t, app,
		fmt.Sprintf("chat_user_2_%s@example.com", uniqueSuffix),
		"password456",
		"ChatReceiver", "Two")

	t.Run("SendMessage - Success", func(t *testing.T) {
		payload := models.SendMessageM{
			Sender:  user1ID,
			Recever: user2ID,
			Content: "Hello from modular test! How are you doing today?",
		}

		resp, body, err := MakeRequest(app, "POST", "/chat/sendmessage", payload, "")
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected 201 Created, got %d, body: %s", resp.StatusCode, string(body))
		}

		var res struct {
			Message string      `json:"message"`
			Result  interface{} `json:"result"`
		}
		if err := json.Unmarshal(body, &res); err != nil {
			t.Fatalf("Failed to parse send message response: %v", err)
		}
		if res.Result == nil {
			t.Error("Expected non-nil inserted message ID")
		}
	})

	t.Run("GetMsgsByNums - Retrieve Message History", func(t *testing.T) {
		url := fmt.Sprintf("/chat/getmsgsbynums?from=0&firstuid=%s&seconduid=%s", user1ID, user2ID)
		resp, body, err := MakeRequest(app, "GET", url, nil, "")
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d, body: %s", resp.StatusCode, string(body))
		}

		var res struct {
			Msgs []models.Message `json:"msgs"`
		}
		if err := json.Unmarshal(body, &res); err != nil {
			t.Fatalf("Failed to parse messages response: %v", err)
		}
		if len(res.Msgs) == 0 {
			t.Error("Expected at least 1 message in chat history")
		}
	})

	t.Run("GetUserUnreadedMsg - Verify Unread Count for Receiver", func(t *testing.T) {
		resp, body, err := MakeRequest(app, "GET", "/chat/get-user-unreadedmsg?userid="+user2ID, nil, "")
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d, body: %s", resp.StatusCode, string(body))
		}

		var res struct {
			Messages []models.UnReadedMsg `json:"messages"`
			Total    int                  `json:"total"`
		}
		if err := json.Unmarshal(body, &res); err != nil {
			t.Fatalf("Failed to parse unread messages response: %v", err)
		}
		if res.Total < 1 {
			t.Errorf("Expected at least 1 unread message, got %d", res.Total)
		}
	})

	t.Run("MarkMsgAsReaded - Mark as Read", func(t *testing.T) {
		url := fmt.Sprintf("/chat/mark-msg-asreaded?mainuid=%s&otheruid=%s", user2ID, user1ID)
		resp, body, err := MakeRequest(app, "GET", url, nil, "")
		if err != nil {
			t.Fatalf("Unexpected request error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d, body: %s", resp.StatusCode, string(body))
		}

		var res struct {
			IsMarked bool `json:"isMarked"`
		}
		if err := json.Unmarshal(body, &res); err != nil {
			t.Fatalf("Failed to parse mark response: %v", err)
		}
		if !res.IsMarked {
			t.Error("Expected isMarked to be true")
		}
	})
}
