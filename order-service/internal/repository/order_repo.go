package repository

import (
	"context"
	"encoding/json"
	"order_service/internal/model"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type OrderRepo struct {
	db *sqlx.DB
}

func NewOrderRepo(db *sqlx.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}

// Прописали значит метод который сохраняет в бд наш EventOrder
func (r *OrderRepo) SaveTx(ctx context.Context, tx *sqlx.Tx, order model.Order) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO orders (id, user_id, item, status, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		order.Id,
		order.UserId,
		order.Item,
		order.Status,
		order.CreatedAt,
	)
	return err
}

func (r *OrderRepo) SaveOutboxTx(ctx context.Context, tx *sqlx.Tx, order model.Order) error {

	//Маршалим структуру в джсон - чтобы воркер смог забрать ее в кафку
	payload, err := json.Marshal(model.EventOrder{
		EventType: "order.created",
		Order:     order,
	})
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO outbox (id, order_id, event_type, payload, sent, created_at)
		 VALUES ($1, $2, $3, $4, false, $5)`,
		uuid.New().String(),
		order.Id,
		"order.created",
		string(payload),
		time.Now(),
	)
	return err
}

// Тут обрати внимание у нас метод который все собирает, он рождает транзакцию и по очереди прокидывает ее вметоды
// SaveTx и SaveOutboxTx если ошибка - будем откатываться (defer rollback), если нет - то коммит
func (r *OrderRepo) SaveWithOutbox(ctx context.Context, order model.Order) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// INSERT 1 — сохраняем заказ
	if err := r.SaveTx(ctx, tx, order); err != nil {
		return err
	}

	// INSERT 2 — сохраняем событие в outbox
	if err := r.SaveOutboxTx(ctx, tx, order); err != nil {
		return err
	}

	// оба INSERT фиксируются атомарно
	return tx.Commit()
}

// читаем неотправленные события
func (r *OrderRepo) GetUnsentEvents(ctx context.Context) ([]model.OutboxEvent, error) {
	var events []model.OutboxEvent
	err := r.db.SelectContext(ctx, &events,
		`SELECT * FROM outbox WHERE sent = false ORDER BY created_at ASC`)
	return events, err
}

// помечаем событие как отправленное
func (r *OrderRepo) MarkAsSent(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE outbox SET sent = true WHERE id = $1`, id)
	return err
}
