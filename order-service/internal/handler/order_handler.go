package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"order_service/internal/model"
	"order_service/internal/repository"
	"time"

	"github.com/google/uuid"
)

type OrderHandler struct {
	repo *repository.OrderRepo
}

func NewOrderHandler(s *repository.OrderRepo) *OrderHandler {
	return &OrderHandler{repo: s}
}

func (s *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {

	//попробуем вычитаь в order тело нашего запроса
	var order model.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	//Заполним служебный поля которые не передает клиент
	order.Id = uuid.New().String()
	order.Status = "сreated"
	order.CreatedAt = time.Now()

	if err := s.repo.SaveWithOutbox(r.Context(), order); err != nil {
		log.Printf("failed to save order: %v", err)
		http.Error(w, "internal service error", http.StatusInternalServerError)
	}

	log.Printf("order created: %s", order.Id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)

}
