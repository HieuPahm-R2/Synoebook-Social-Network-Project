package realtime

import (
	"log"
	"realTimeChat/grpc"
)

func GetUserFriends(userID string) <-chan []string {
	// Unbuffered channel leak có thể gây leak goroutine nếu phía nhận dừng lắng nghe
	ch := make(chan []string, 1)
	go func() {
		defer close(ch)

		// call grpc client func
		userFriends, err := grpc.GetFollowingFollowersClient(userID)
		if err != nil {
			log.Printf("Error getting friends for user %s: %v", userID, err)
			ch <- []string{}
			return
		}

		seen := make(map[string]bool)
		var friends []string
		for _, userIDsList := range userFriends {
			if userIDsList == nil {
				continue
			}
			for _, id := range userIDsList.UserIdsList {
				if id != "" && id != userID && !seen[id] {
					seen[id] = true
					friends = append(friends, id)
				}
			}
		}

		ch <- friends
	}()
	return ch
}

// func GetUserFriends(userID string) <-chan []string {
// 	ch := make(chan []string)
// 	go func() {
// 		defer close(ch)
// 		switch userID {
// 		case "1":
// 			ch <- []string{"2", "3", "4"}
// 		case "2":
// 			ch <- []string{"1", "3", "4"}
// 		case "3":
// 			ch <- []string{"1", "2", "4"}
// 		case "4":
// 			ch <- []string{"1"}
// 		default:
// 			ch <- []string{}
// 		}
// 	}()
// 	return ch

// }
