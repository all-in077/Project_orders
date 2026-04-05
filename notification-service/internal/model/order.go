package model

import "time"

// Опишем тип запроса для json заказа
type Order struct {
	Id     string `json:"id"`
	UserId string `json:"user_id"`
	Item   string `json:"item"`
	Status string `json:"status"`
}

// Опишем событие которое будет приходить в кафку
type EventOrder struct {
	EventType string `json:"event_type"`
	Order     Order  `json:"order"` //тут у нас вложенный джосн Order со всеми полями
}

// Напишем структру которую будем сохранять в бд исходя из сообщения
type Notification struct {
	Id        string    `db:"id"`
	OrderId   string    `db:"order_id"`
	UserId    string    `db:"user_id"`
	Message   string    `db:"message"`
	CreatedAt time.Time `db:"created_at"`
}
