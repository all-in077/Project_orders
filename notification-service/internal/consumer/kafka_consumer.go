package consumer

import (
	"context"
	"encoding/json"
	"log"
	"notification-service/internal/model"
	"notification-service/internal/notifier"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader   *kafka.Reader
	notifier *notifier.Notifier
}

// Соберем конструктор для consumer-а тут мы передали адрес брокера, топик и
func NewKafkaConsumer(addr string, topic string, groupId string, n *notifier.Notifier) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{addr},
		Topic:   topic,
		GroupID: groupId, //Обозначение для единной группу консьюмеров
	})

	return &KafkaConsumer{
		reader:   reader,
		notifier: n,
	}
}

func (s *KafkaConsumer) Start(ctx context.Context) {
	log.Println("consumer started, waiting for messages...")

	for {
		msg, err := s.reader.FetchMessage(ctx) //FetchMessage читает сообщение  из кафки но не комитит
		if err != nil {
			if ctx.Err() != nil {
				log.Println("context cancelled, stopping consumer")
				return
			}
			log.Printf("fetch error: %v", err)
			continue
		}
		var event model.EventOrder
		if err := json.Unmarshal(msg.Value, &event); err != nil { //То что мы прочитали, возьмем от структруы поле Value - слайс байт с нашим json
			log.Printf("faildef to unmarshal message: %v", err)
			s.reader.CommitMessages(ctx, msg)
			continue
		}

		s.notifier.Notify(ctx, event) //Попробуем узнать сообщение и записать его в бд
		if err := s.reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("failed to commot message: %v", err)
		}

	}

}

func (s *KafkaConsumer) Close() error {
	return s.reader.Close()
}
