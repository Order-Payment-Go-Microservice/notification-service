package service

import (
	"testing"

	"github.com/Order-Payment-Go-Microservice/notification-service/internal/model"
	"github.com/google/uuid"
)

type mockNotificationRepo struct {
	items []model.Notification
}

func (m *mockNotificationRepo) Create(n *model.Notification) error {
	m.items = append(m.items, *n)
	return nil
}
func (m *mockNotificationRepo) GetByID(id uuid.UUID) (*model.Notification, error) {
	for _, n := range m.items {
		if n.ID == id {
			return &n, nil
		}
	}
	return nil, nil
}
func (m *mockNotificationRepo) GetByUserID(userID uuid.UUID) ([]model.Notification, error) {
	return m.items, nil
}
func (m *mockNotificationRepo) Delete(id uuid.UUID) error { return nil }
func (m *mockNotificationRepo) MarkRead(id uuid.UUID) error {
	return nil
}
func (m *mockNotificationRepo) MarkAllRead(userID uuid.UUID) error { return nil }
func (m *mockNotificationRepo) CountUnread(userID uuid.UUID) (int, error) { return 0, nil }
func (m *mockNotificationRepo) Search(userID uuid.UUID, query string) ([]model.Notification, error) {
	return nil, nil
}

func TestCreateNotification(t *testing.T) {
	repo := &mockNotificationRepo{}
	svc := NewNotificationService(repo, nil)

	userID := uuid.New()
	n, err := svc.CreateNotification(userID, "Test Title", "Test Message", "push")
	if err != nil {
		t.Fatalf("CreateNotification failed: %v", err)
	}
	if n.Title != "Test Title" {
		t.Errorf("expected Test Title, got %s", n.Title)
	}
	if n.CreatedAt.IsZero() {
		t.Error("created_at should be set")
	}
}
