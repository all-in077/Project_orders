package repository

import (
	"context"
	"notification-service/internal/model"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type NotificationRepo struct {
	db *sqlx.DB
}

func NewNotificationRepository(db *sqlx.DB) *NotificationRepo {
	return &NotificationRepo{db: db}
}

// Прописали значит метод который сохраняет в бд наш EventOrder
func (s *NotificationRepo) Save(ctx context.Context, event model.EventOrder) error {

	notification := &model.Notification{
		Id:        uuid.New().String(),
		OrderId:   event.Order.Id,
		UserId:    event.Order.UserId,
		Message:   "ваш заказ: " + event.Order.Id + "принят в обработку",
		CreatedAt: time.Now(),
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO notifications (id, order_id, user_id, message, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		notification.Id,
		notification.OrderId,
		notification.UserId,
		notification.Message,
		notification.CreatedAt,
	)
	return err

}
