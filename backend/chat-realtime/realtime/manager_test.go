package realtime

import (
	"testing"
	"time"
)

func TestConnectionManager_AddAndRemove(t *testing.T) {
	mockFriends := func(id string) <-chan []string {
		ch := make(chan []string, 1)
		defer close(ch)
		if id == "u1" {
			ch <- []string{"u2"}
		} else if id == "u2" {
			ch <- []string{"u1"}
		} else {
			ch <- []string{}
		}
		return ch
	}

	cm := NewConnectionManager(mockFriends)
	if cm == nil {
		t.Fatal("expected ConnectionManager to be non-nil")
	}

	// Verify isFriend
	if !cm.isFriend("u1", "u2") {
		t.Errorf("expected u1 and u2 to be friends")
	}
	if cm.isFriend("u1", "u3") {
		t.Errorf("expected u1 and u3 to NOT be friends")
	}
}

func TestConnectionManager_OfflineMessage(t *testing.T) {
	cm := NewConnectionManager(nil)
	msg := Message{
		Sender:  "u1",
		Recever: "u2",
		Content: "hello",
	}

	// Sending to offline user should not panic or block
	done := make(chan struct{})
	go func() {
		cm.SendToReceiver(msg)
		close(done)
	}()

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("SendToReceiver timed out")
	}
}
