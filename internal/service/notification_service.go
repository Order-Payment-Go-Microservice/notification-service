package service

import (
	"context"
	"log"
	"time"

	"github.com/Order-Payment-Go-Microservice/notification-service/internal/model"
	"github.com/Order-Payment-Go-Microservice/notification-service/internal/repository"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type NotificationService interface {
	CreateNotification(userID uuid.UUID, title, message, nType string) (*model.Notification, error)
	GetHistory(userID uuid.UUID) ([]model.Notification, error)
	Delete(id uuid.UUID) error
	MarkRead(id uuid.UUID) error
	MarkAllRead(userID uuid.UUID) error
	GetUnreadCount(userID uuid.UUID) (int, error)
	Search(userID uuid.UUID, query string) ([]model.Notification, error)
}

type notificationService struct {
	repo        repository.NotificationRepository
	redisClient *redis.Client
}

func NewNotificationService(repo repository.NotificationRepository, rdb *redis.Client) NotificationService {
	return &notificationService{repo: repo, redisClient: rdb}
}

func (s *notificationService) CreateNotification(userID uuid.UUID, title, message, nType string) (*model.Notification, error) {
	notification := &model.Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     title,
		Message:   message,
		Type:      nType,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(notification); err != nil {
		return nil, err
	}

	s.invalidateUnreadCache(userID)
	return notification, nil
}

func (s *notificationService) GetHistory(userID uuid.UUID) ([]model.Notification, error) {
	return s.repo.GetByUserID(userID)
}

func (s *notificationService) Delete(id uuid.UUID) error {
	n, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	s.invalidateUnreadCache(n.UserID)
	return nil
}

func (s *notificationService) MarkRead(id uuid.UUID) error {
	n, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if err := s.repo.MarkRead(id); err != nil {
		return err
	}
	s.invalidateUnreadCache(n.UserID)
	return nil
}

func (s *notificationService) MarkAllRead(userID uuid.UUID) error {
	if err := s.repo.MarkAllRead(userID); err != nil {
		return err
	}
	s.invalidateUnreadCache(userID)
	return nil
}

func (s *notificationService) GetUnreadCount(userID uuid.UUID) (int, error) {
	ctx := context.Background()
	cacheKey := "unread_count:" + userID.String()

	if s.redisClient != nil {
		cached, err := s.redisClient.Get(ctx, cacheKey).Int()
		if err == nil {
			log.Println("[Redis] Unread count loaded from cache")
			return cached, nil
		}
	}

	count, err := s.repo.CountUnread(userID)
	if err != nil {
		return 0, err
	}

	if s.redisClient != nil {
		s.redisClient.Set(ctx, cacheKey, count, 10*time.Minute)
	}
	return count, nil
}

func (s *notificationService) Search(userID uuid.UUID, query string) ([]model.Notification, error) {
	return s.repo.Search(userID, query)
}

func (s *notificationService) invalidateUnreadCache(userID uuid.UUID) {
	if s.redisClient == nil {
		return
	}
	ctx := context.Background()
	s.redisClient.Del(ctx, "unread_count:"+userID.String())
}
