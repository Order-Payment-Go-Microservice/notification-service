package repository

import (
	"database/sql"
	"fmt"

	"github.com/Order-Payment-Go-Microservice/notification-service/internal/model"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type NotificationRepository interface {
	Create(n *model.Notification) error
	GetByID(id uuid.UUID) (*model.Notification, error)
	GetByUserID(userID uuid.UUID) ([]model.Notification, error)
	Delete(id uuid.UUID) error
	MarkRead(id uuid.UUID) error
	MarkAllRead(userID uuid.UUID) error
	CountUnread(userID uuid.UUID) (int, error)
	Search(userID uuid.UUID, query string) ([]model.Notification, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) NotificationRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(n *model.Notification) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `INSERT INTO notifications (id, user_id, title, message, type, is_read, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	if _, err = tx.Exec(query, n.ID, n.UserID, n.Title, n.Message, n.Type, n.IsRead, n.CreatedAt); err != nil {
		return fmt.Errorf("insert notification: %w", err)
	}
	return tx.Commit()
}

func (r *postgresRepository) GetByID(id uuid.UUID) (*model.Notification, error) {
	query := `SELECT id, user_id, title, message, type, is_read, created_at FROM notifications WHERE id = $1`
	row := r.db.QueryRow(query, id)
	var n model.Notification
	if err := row.Scan(&n.ID, &n.UserID, &n.Title, &n.Message, &n.Type, &n.IsRead, &n.CreatedAt); err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *postgresRepository) GetByUserID(userID uuid.UUID) ([]model.Notification, error) {
	query := `SELECT id, user_id, title, message, type, is_read, created_at FROM notifications WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []model.Notification
	for rows.Next() {
		var n model.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Message, &n.Type, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}

func (r *postgresRepository) Delete(id uuid.UUID) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec(`DELETE FROM notifications WHERE id = $1`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *postgresRepository) MarkRead(id uuid.UUID) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec(`UPDATE notifications SET is_read = true WHERE id = $1`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *postgresRepository) MarkAllRead(userID uuid.UUID) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec(`UPDATE notifications SET is_read = true WHERE user_id = $1`, userID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *postgresRepository) CountUnread(userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`
	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

func (r *postgresRepository) Search(userID uuid.UUID, query string) ([]model.Notification, error) {
	sqlQuery := `SELECT id, user_id, title, message, type, is_read, created_at 
                 FROM notifications WHERE user_id = $1 AND (title ILIKE '%' || $2 || '%' OR message ILIKE '%' || $2 || '%')
                 ORDER BY created_at DESC`
	rows, err := r.db.Query(sqlQuery, userID, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []model.Notification
	for rows.Next() {
		var n model.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Message, &n.Type, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}
