package notifier

import (
	"context"
	"log"
	"notification-service/internal/model"
	"notification-service/internal/repository"
)

type Notifier struct {
	repo *repository.NotificationRepo
}

func NewNotifier(repo *repository.NotificationRepo) *Notifier {
	return &Notifier{repo: repo}
}

// Добавили в метод Notify запись в бд путем встраивания в структуру Notifier - repo который позволяет писать в бд
func (n *Notifier) Notify(ctx context.Context, event model.EventOrder) {
	switch event.EventType {
	case "order.created":
		if err := n.repo.Save(ctx, event); err != nil {
			log.Printf("failed to save notivication: %v", err)
			return
		}
		log.Printf("notification saved for order #%s user %s", event.Order.Id, event.Order.UserId)
	default:
		log.Printf("unknown eventType: %s", event.EventType)
	}
}
