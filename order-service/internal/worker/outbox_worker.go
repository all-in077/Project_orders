package worker

import (
	"context"
	"log"
	"order_service/internal/producer"
	"order_service/internal/repository"
	"time"
)

type OutboxWorker struct {
	repo     *repository.OrderRepo
	producer *producer.KafkaProducer
}

func NewOutboxWorker(repo *repository.OrderRepo, producer *producer.KafkaProducer) *OutboxWorker {
	return &OutboxWorker{
		repo:     repo,
		producer: producer,
	}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	log.Println("outbox worker started")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("outbox worker stopping")
			return
		case <-ticker.C:
			w.processOutbox(ctx)
		}
	}
}

func (w *OutboxWorker) processOutbox(ctx context.Context) {
	events, err := w.repo.GetUnsentEvents(ctx) //Получаем все ивенты которые не были отправлены в кафку ( не были отмечены в поле sent)
	if err != nil {
		log.Printf("failed to get unsent events: %v", err)
		return
	}

	if len(events) == 0 {
		return
	}

	log.Printf("found %d unsent events", len(events))

	for _, event := range events {
		if err := w.producer.PublishRaw(ctx, event.Payload); err != nil {
			log.Printf("failed to publish event %s: %v", event.Id, err)
			continue // не помечаем как sent — попробуем в следующий раз
		}

		if err := w.repo.MarkAsSent(ctx, event.Id); err != nil {
			log.Printf("failed to mark event as sent %s: %v", event.Id, err)
		}

		log.Printf("event sent: %s", event.Id)
	}
}
