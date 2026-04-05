package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"order_service/internal/model"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(port string, topic string) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(port),
		Topic:        topic,
		Balancer:     &kafka.Hash{},
		BatchTimeout: 10 * time.Microsecond,
		RequiredAcks: kafka.RequireOne,
	}
	return &KafkaProducer{writer: writer}
}

// Передали в метод контекст - для вынужденной отмены в случае чего и сам event который формируем в kafka сообщение (через marshal -> json) и пытемся записать его в топик
func (s *KafkaProducer) PublishOrderEvent(ctx context.Context, event model.EventOrder) error {

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err) // Обарачиваем и возвращаем ошибку
	}

	msg := kafka.Message{
		Key:   []byte(event.Order.Id),
		Value: data,
	}

	if err := s.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to write message %w", err) //Оборачиваем и возвращаем ошибку
	}

	return nil
}

func (s *KafkaProducer) PublishRaw(ctx context.Context, payload string) error {
	msg := kafka.Message{
		Value: []byte(payload),
	}

	if err := s.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

func (s *KafkaProducer) Close() error {
	return s.writer.Close()
}
