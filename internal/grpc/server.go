package grpc

import (
	"context"
	"log"
	"notification-service/internal/service"
	pb "notification-service/proto"

	"github.com/google/uuid"
)

type NotificationServer struct {
	pb.UnimplementedNotificationServiceServer
	svc service.NotificationService
}

func NewNotificationServer(svc service.NotificationService) *NotificationServer {
	return &NotificationServer{svc: svc}
}

func (s *NotificationServer) SendNotification(ctx context.Context, req *pb.NotificationRequest) (*pb.NotificationResponse, error) {
	log.Printf("[gRPC Server] Received notification request for user: %s", req.GetUserId())
	
	uID, _ := uuid.Parse(req.GetUserId())
	_, err := s.svc.CreateNotification(uID, req.GetTitle(), req.GetMessage(), "push")
	if err != nil {
		return nil, err
	}

	return &pb.NotificationResponse{Status: "sent"}, nil
}
