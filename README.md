# Orders Platform

Микросервисная система обработки заказов на Go.

## Сервисы

- **order-service** — создание заказов, сохранение в PostgreSQL, отправка событий в Kafka
- **notification-service** — получение событий из Kafka, сохранение уведомлений в PostgreSQL с информацие по заказу

## Стек

- Go 1.24
- PostgreSQL 15
- Apache Kafka
- Docker / Docker Compose

## Запуск
```bash
git clone https://github.com/all-in077/Project_orders
cd Project_orders
docker-compose up --build
```

## API

### Orders

POST /orders   Создать заказ 