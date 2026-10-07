package grpc

import (
	"context"
	"log"
	"os"
	"realTimeChat/protos"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func getGRPCAddr() string {
	addr := os.Getenv("GRPC_SERVER_ADDR")
	if addr == "" {
		addr = os.Getenv("GRPC_URL")
	}
	if addr == "" {
		addr = ":5001"
	}
	return addr
}

func GetFollowingFollowersClient(id string) ([]*protos.UserIDsList, error) {
	conn, err := grpc.NewClient(getGRPCAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("failed to connect to gRPC server: %v", err)
		return nil, err
	}
	defer conn.Close()

	client := protos.NewRealtimeChatServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &protos.UserID{Userid: id}
	resp, err := client.GetUserFollowingFollowers(ctx, req)
	if err != nil {
		//log.Fatalf không lưu được tin nhắn qua gRPC. Nếu lỗi mạng, log.Fatalf sẽ gọi os.Exit(1) làm sập toàn bộ WebSocket server.
		log.Printf("error calling GetUserFollowingFollowers gRPC: %v", err)
		return nil, err
	}

	if resp == nil {
		return nil, nil
	}

	return resp.GetUserIDsLists(), nil
}

func SendMessageClient(sender, receiver, content string) error {
	conn, err := grpc.NewClient(getGRPCAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("failed to connect to gRPC server: %v", err)
		return err
	}
	defer conn.Close()

	client := protos.NewRealtimeChatServiceClient(conn)
	// Dùng context.Background() khiến request gRPC có thể treo nếu server không phản hồi.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &protos.MessageRequest{
		Sender:   sender,
		Receiver: receiver,
		Content:  content,
	}

	_, err = client.SendMessage(ctx, req)
	if err != nil {
		log.Printf("error calling SendMessage gRPC: %v", err)
		return err
	}

	return nil
}
