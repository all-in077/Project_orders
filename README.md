## Сервисы

- **api-gateway** - точка входа, JWT авторизация через gRPC (общение с auth-service), reverse proxy для маршрутизации запросов на нуные сервисы
- **auth-service** - регистрация и вход (JWT токены)
- **order-service** - создание заказов, сохранение в PostgreSQL, отправка событий в Kafka
- **notification-service** - Kafka consumer, обработка событий заказов, сохранение уведомлений

## Стек

- Go 1.24
- PostgreSQL 15
- Apache Kafka
- Docker / Docker Compose
- gRPC + Protocol Buffers
- JWT (golang-jwt/jwt)

## Запуск

```bash
git clone https://github.com/all-in077/Project_orders
cd Project_orders
docker-compose up --build
```

## API

Все запросы идут через api-gateway на порту `:8080`

### Auth
POST /auth/register Регистрация нового пользователя |
POST /auth/login  Вход, возвращает access и refresh токены |
POST /auth/refresh  Обновление access токена 

### Orders 
POST /orders Создать заказ 