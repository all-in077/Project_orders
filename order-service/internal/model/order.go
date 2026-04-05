package model

import "time"

// Опишем тип запроса для json заказа
type Order struct {
	Id        string    `json:"id" db:"id"`
	UserId    string    `json:"user_id" db:"user_id"`
	Item      string    `json:"item" db:"item"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Опишем событие которое будет приходить в кафку
type EventOrder struct {
	EventType string `json:"event_type"`
	Order     Order  `json:"order"` //тут у нас вложенный джосн Order со всеми полями
}

type OutboxEvent struct {
	Id        string    `db:"id"`
	OrderId   string    `db:"order_id"`
	EventType string    `db:"event_type"`
	Payload   string    `db:"payload"`
	Sent      bool      `db:"sent"`
	CreatedAt time.Time `db:"created_at"`
}
