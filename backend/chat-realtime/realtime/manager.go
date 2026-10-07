package realtime

import (
	"log"
	"realTimeChat/grpc"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type Message struct {
	Sender  string `json:"sender"`
	Recever string `json:"recever"`
	Content string `json:"content"`
}

type ConnectionManager struct {
	connections    map[string]*websocket.Conn
	onlineFriends  map[string][]string
	getUserFriends func(string) <-chan []string
	lock           sync.RWMutex
}

func NewConnectionManager(getUserFriends func(string) <-chan []string) *ConnectionManager {
	if getUserFriends == nil {
		getUserFriends = GetUserFriends
	}
	return &ConnectionManager{
		connections:    make(map[string]*websocket.Conn),
		onlineFriends:  make(map[string][]string),
		getUserFriends: getUserFriends,
	}
}

// map connections và slice onlineFriends bị đọc/ghi mà không có lock, gây lỗi fatal error: concurrent map writes
func (cm *ConnectionManager) AddConnection(userID string, conn *websocket.Conn) {
	cm.lock.Lock()
	cm.connections[userID] = conn
	cm.onlineFriends[userID] = []string{}
	cm.lock.Unlock()

	// Update online friends asynchronously to avoid blocking connection handling
	go func() {
		// 1. Fetch friend IDs from gRPC service
		var friendIDs []string
		for friends := range cm.getUserFriends(userID) {
			if friends != nil {
				friendIDs = append(friendIDs, friends...)
			}
		}

		friendSet := make(map[string]bool)
		for _, f := range friendIDs {
			friendSet[f] = true
		}

		type notification struct {
			conn    *websocket.Conn
			friends []string
		}
		var toNotify []notification

		cm.lock.Lock()
		// Verify this connection is still active
		currentConn, active := cm.connections[userID]
		if !active || currentConn != conn {
			cm.lock.Unlock()
			return
		}

		var currentOnlineFriends []string

		// Check which friends are online and notify them
		for friendID := range cm.connections {
			if friendID == userID {
				continue
			}
			if friendSet[friendID] {
				currentOnlineFriends = append(currentOnlineFriends, friendID)

				// Add userID to friendID's onlineFriends list if not already present
				alreadyAdded := false
				for _, id := range cm.onlineFriends[friendID] {
					if id == userID {
						alreadyAdded = true
						break
					}
				}
				if !alreadyAdded {
					cm.onlineFriends[friendID] = append(cm.onlineFriends[friendID], userID)
				}

				if friendConn := cm.connections[friendID]; friendConn != nil {
					toNotify = append(toNotify, notification{
						conn:    friendConn,
						friends: append([]string(nil), cm.onlineFriends[friendID]...),
					})
				}
			}
		}

		cm.onlineFriends[userID] = currentOnlineFriends
		userFriendsToSend := append([]string(nil), currentOnlineFriends...)
		cm.lock.Unlock()

		// Send notifications outside the lock
		for _, n := range toNotify {
			if err := n.conn.WriteJSON(map[string]interface{}{
				"onlineFriends": n.friends,
			}); err != nil {
				log.Printf("Error notifying friend about online user %s: %v", userID, err)
			}
		}

		// Notify newly connected user with all their online friends
		if err := conn.WriteJSON(map[string]interface{}{
			"onlineFriends": userFriendsToSend,
		}); err != nil {
			log.Printf("Error notifying %s about online friends: %v", userID, err)
		}
	}()
}

func (cm *ConnectionManager) RemoveConnection(userID string) {
	type notification struct {
		conn    *websocket.Conn
		friends []string
	}
	var toNotify []notification

	cm.lock.Lock()
	delete(cm.connections, userID)
	delete(cm.onlineFriends, userID)

	for friendID := range cm.onlineFriends {
		for i, id := range cm.onlineFriends[friendID] {
			if id == userID {
				cm.onlineFriends[friendID] = append(cm.onlineFriends[friendID][:i], cm.onlineFriends[friendID][i+1:]...)
				if friendConn, ok := cm.connections[friendID]; ok && friendConn != nil {
					toNotify = append(toNotify, notification{
						conn:    friendConn,
						friends: append([]string(nil), cm.onlineFriends[friendID]...),
					})
				}
				break
			}
		}
	}
	cm.lock.Unlock()

	// Notify online friends outside the lock
	for _, n := range toNotify {
		if err := n.conn.WriteJSON(map[string]interface{}{
			"onlineFriends": n.friends,
		}); err != nil {
			log.Printf("Error notifying friend about disconnect of %s: %v", userID, err)
		}
	}
}

func (cm *ConnectionManager) SendToReceiver(msg Message) {
	var targetConn *websocket.Conn

	cm.lock.RLock()
	if conn, ok := cm.connections[msg.Recever]; ok {
		targetConn = conn
	}
	cm.lock.RUnlock()

	// If receiver is connected, deliver message via WebSocket
	if targetConn != nil {
		if err := targetConn.WriteJSON(msg); err != nil {
			log.Printf("Error sending message to %s: %v", msg.Recever, err)
		}
	} else {
		log.Printf("Receiver %s is not currently online", msg.Recever)
	}

	// Always persist message to DB via gRPC (even if receiver is offline)
	go func() {
		if err := grpc.SendMessageClient(msg.Sender, msg.Recever, msg.Content); err != nil {
			log.Printf("Error saving message via gRPC for receiver %s: %v", msg.Recever, err)
		}
	}()
}

// helper func isFriend
func (cm *ConnectionManager) isFriend(userID, friendID string) bool {
	friends := <-cm.getUserFriends(userID)
	if friends == nil {
		return false
	}
	for _, f := range friends {
		if f == friendID {
			return true
		}
	}
	return false
}

// StartWebSocketServer starts Fiber HTTP & WebSocket server
func StartWebSocketServer(addr string, cm *ConnectionManager) error {
	app := fiber.New()

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		userID := c.Params("id")
		if userID == "" {
			return
		}

		cm.AddConnection(userID, c)
		defer func() {
			cm.RemoveConnection(userID)
			c.Close()
		}()

		for {
			var msg Message
			if err := c.ReadJSON(&msg); err != nil {
				break
			}
			if msg.Sender == "" {
				msg.Sender = userID
			}
			cm.SendToReceiver(msg)
		}
	}))

	return app.Listen(addr)
}
