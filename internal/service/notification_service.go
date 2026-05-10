package service

import (
	"log"
	"notification-service/internal/model"
	"notification-service/internal/repository"
	"time"

	"github.com/google/uuid"
)

type NotificationService interface {
	CreateNotification(userID uuid.UUID, title, message, notifType string) (*model.Notification, error)
	GetUserNotifications(userID string) ([]*model.Notification, error)
	MarkRead(id string) error
}

type notificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) NotificationService {
	return &notificationService{repo: repo}
}

func (s *notificationService) CreateNotification(userID uuid.UUID, title, message, notifType string) (*model.Notification, error) {
	n := &model.Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     title,
		Message:   message,
		Type:      notifType,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Save(n); err != nil {
		return nil, err
	}

	log.Printf("[Notification Service] %s dispatched to User %s: %s - %s", n.Type, n.UserID, n.Title, n.Message)
	if n.Type == "email" {
		log.Printf("[Email simulation] Sending email to user %s...", n.UserID)
	}

	return n, nil
}

func (s *notificationService) GetUserNotifications(userIDStr string) ([]*model.Notification, error) {
	userID, _ := uuid.Parse(userIDStr)
	return s.repo.GetByUserID(userID)
}

func (s *notificationService) MarkRead(idStr string) error {
	id, _ := uuid.Parse(idStr)
	return s.repo.MarkAsRead(id)
}
