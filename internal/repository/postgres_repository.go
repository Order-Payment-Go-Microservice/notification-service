package repository

import (
	"database/sql"
	"github.com/Order-Payment-Go-Microservice/notification-service/internal/model"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type NotificationRepository interface {
	Save(n *model.Notification) error
	GetByUserID(userID uuid.UUID) ([]*model.Notification, error)
	MarkAsRead(id uuid.UUID) error
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) NotificationRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Save(n *model.Notification) error {
	query := `INSERT INTO notifications (id, user_id, title, message, type, is_read, created_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(query, n.ID, n.UserID, n.Title, n.Message, n.Type, n.IsRead, n.CreatedAt)
	return err
}

func (r *postgresRepository) GetByUserID(userID uuid.UUID) ([]*model.Notification, error) {
	query := `SELECT id, user_id, title, message, type, is_read, created_at 
              FROM notifications WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*model.Notification
	for rows.Next() {
		n := &model.Notification{}
		err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Message, &n.Type, &n.IsRead, &n.CreatedAt)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}

func (r *postgresRepository) MarkAsRead(id uuid.UUID) error {
	query := `UPDATE notifications SET is_read = true WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
