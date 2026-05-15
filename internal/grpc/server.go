package grpc

import (
	"context"

	"github.com/Order-Payment-Go-Microservice/notification-service/internal/service"
	pb "github.com/Order-Payment-Go-Microservice/proto-generation/gen/notification/v1"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

type NotificationServer struct {
	pb.UnimplementedNotificationServiceServer
	svc service.NotificationService
}

func NewNotificationServer(svc service.NotificationService) *NotificationServer {
	return &NotificationServer{svc: svc}
}

func (s *NotificationServer) SendNotification(ctx context.Context, req *pb.NotificationRequest) (*pb.NotificationResponse, error) {
	uID, _ := uuid.Parse(req.UserId)
	n, err := s.svc.CreateNotification(uID, req.Title, req.Message, req.Type)
	if err != nil {
		return nil, err
	}
	return &pb.NotificationResponse{Id: n.ID.String(), Status: "sent"}, nil
}

func (s *NotificationServer) GetNotificationHistory(ctx context.Context, req *pb.GetHistoryRequest) (*pb.GetHistoryResponse, error) {
	uID, _ := uuid.Parse(req.UserId)
	notifications, err := s.svc.GetHistory(uID)
	if err != nil {
		return nil, err
	}
	list := make([]*pb.NotificationDetail, len(notifications))
	for i, n := range notifications {
		list[i] = &pb.NotificationDetail{
			Id: n.ID.String(), Title: n.Title, Message: n.Message,
			IsRead: n.IsRead, CreatedAt: n.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	return &pb.GetHistoryResponse{Notifications: list}, nil
}

func (s *NotificationServer) DeleteNotification(ctx context.Context, req *pb.DeleteRequest) (*emptypb.Empty, error) {
	id, _ := uuid.Parse(req.Id)
	if err := s.svc.Delete(id); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *NotificationServer) MarkNotificationRead(ctx context.Context, req *pb.MarkReadRequest) (*emptypb.Empty, error) {
	id, _ := uuid.Parse(req.Id)
	if err := s.svc.MarkRead(id); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *NotificationServer) GetUnreadCount(ctx context.Context, req *pb.GetUnreadCountRequest) (*pb.UnreadCountResponse, error) {
	uID, _ := uuid.Parse(req.UserId)
	count, err := s.svc.GetUnreadCount(uID)
	if err != nil {
		return nil, err
	}
	return &pb.UnreadCountResponse{Count: int32(count)}, nil
}

func (s *NotificationServer) MarkAllAsRead(ctx context.Context, req *pb.MarkAllReadRequest) (*emptypb.Empty, error) {
	uID, _ := uuid.Parse(req.UserId)
	if err := s.svc.MarkAllRead(uID); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *NotificationServer) UpdatePreferences(ctx context.Context, req *pb.UpdatePreferencesRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *NotificationServer) GetPreferences(ctx context.Context, req *pb.GetPreferencesRequest) (*pb.PreferencesResponse, error) {
	return &pb.PreferencesResponse{EnableEmail: true, EnablePush: true}, nil
}

func (s *NotificationServer) DeleteAllNotifications(ctx context.Context, req *pb.DeleteAllRequest) (*emptypb.Empty, error) {
	uID, _ := uuid.Parse(req.UserId)
	notifs, _ := s.svc.GetHistory(uID)
	for _, n := range notifs {
		_ = s.svc.Delete(n.ID)
	}
	return &emptypb.Empty{}, nil
}

func (s *NotificationServer) SearchNotifications(ctx context.Context, req *pb.SearchRequest) (*pb.GetHistoryResponse, error) {
	uID, _ := uuid.Parse(req.UserId)
	notifications, err := s.svc.Search(uID, req.Query)
	if err != nil {
		return nil, err
	}
	list := make([]*pb.NotificationDetail, len(notifications))
	for i, n := range notifications {
		list[i] = &pb.NotificationDetail{
			Id: n.ID.String(), Title: n.Title, Message: n.Message,
			IsRead: n.IsRead, CreatedAt: n.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	return &pb.GetHistoryResponse{Notifications: list}, nil
}

func (s *NotificationServer) SubscribeToTopic(ctx context.Context, req *pb.TopicRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *NotificationServer) UnsubscribeFromTopic(ctx context.Context, req *pb.TopicRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
