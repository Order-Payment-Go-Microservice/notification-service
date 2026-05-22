//go:build integration

package repository

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Order-Payment-Go-Microservice/notification-service/internal/model"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func getTestDB(t *testing.T) *sql.DB {
	host := os.Getenv("TEST_DB_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("TEST_DB_PORT")
	if port == "" {
		port = "5435"
	}
	dsn := fmt.Sprintf("host=%s port=%s user=postgres password=postgres dbname=notifications_db sslmode=disable", host, port)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping: %v", err)
	}
	return db
}

func TestIntegration_NotificationCRUD(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	userID := uuid.New()

	notif := &model.Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     "Test Notification",
		Message:   "integration test body",
		Type:      "push",
		IsRead:    false,
		CreatedAt: time.Now().UTC(),
	}
	err := repo.Create(notif)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := repo.GetByID(notif.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Title != "Test Notification" {
		t.Errorf("expected title 'Test Notification', got '%s'", got.Title)
	}
	if got.IsRead {
		t.Error("expected IsRead=false")
	}

	list, err := repo.GetByUserID(userID)
	if err != nil {
		t.Fatalf("GetByUserID failed: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("GetByUserID returned 0 notifications")
	}

	count, err := repo.CountUnread(userID)
	if err != nil {
		t.Fatalf("CountUnread failed: %v", err)
	}
	if count < 1 {
		t.Errorf("expected unread count >= 1, got %d", count)
	}

	err = repo.MarkRead(notif.ID)
	if err != nil {
		t.Fatalf("MarkRead failed: %v", err)
	}
	read, _ := repo.GetByID(notif.ID)
	if !read.IsRead {
		t.Error("expected IsRead=true after MarkRead")
	}

	countAfter, _ := repo.CountUnread(userID)
	if countAfter != count-1 {
		t.Errorf("expected unread count %d, got %d", count-1, countAfter)
	}

	results, err := repo.Search(userID, "integration")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("Search returned 0 results")
	}

	err = repo.Delete(notif.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, err = repo.GetByID(notif.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestIntegration_MarkAllRead(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	userID := uuid.New()

	for i := 0; i < 3; i++ {
		n := &model.Notification{
			ID:        uuid.New(),
			UserID:    userID,
			Title:     fmt.Sprintf("Notif %d", i),
			Message:   "test",
			Type:      "push",
			IsRead:    false,
			CreatedAt: time.Now().UTC(),
		}
		repo.Create(n)
	}

	count, _ := repo.CountUnread(userID)
	if count != 3 {
		t.Fatalf("expected 3 unread, got %d", count)
	}

	err := repo.MarkAllRead(userID)
	if err != nil {
		t.Fatalf("MarkAllRead failed: %v", err)
	}

	countAfter, _ := repo.CountUnread(userID)
	if countAfter != 0 {
		t.Errorf("expected 0 unread after MarkAllRead, got %d", countAfter)
	}

	list, _ := repo.GetByUserID(userID)
	for _, n := range list {
		repo.Delete(n.ID)
	}
}
