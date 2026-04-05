package main

import (
	"context"
	"log"
	"net/http"
	"order_service/internal/handler"
	"order_service/internal/producer"
	"order_service/internal/repository"
	"order_service/internal/worker"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func main() {

	//Подключимся к бд
	db, err := sqlx.Connect("pgx", "postgres://postgres:postgres@localhost:5432/orders_db")
	if err != nil {
		log.Fatal("failed to connect to db:", err)

	}
	defer db.Close()
	log.Println("connected to database")

	//Сосздадаим продюсер на локалхост порт 9092 подпишем топик orders
	kafkaProd := producer.NewKafkaProducer("localhost:9092", "orders")
	defer kafkaProd.Close()

	//Создадим слой с репозиторием который потом передадим в хендлер
	repo := repository.NewOrderRepo(db)
	orderhand := handler.NewOrderHandler(repo)

	//Создадим роутер - маршрутизатор запросов
	router := mux.NewRouter()
	router.HandleFunc("/orders", orderhand.CreateOrder).Methods("POST") //То есть ты буквально говоришь роутеру: "когда придёт POST запрос на /orders — вызови функцию CreateOrder"

	//Создадим сам сервер и передадим ему роутер - которые будет рапсределять запросы
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	ctx, cancel := context.WithCancel(context.Background())

	outboxWorker := worker.NewOutboxWorker(repo, kafkaProd)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		outboxWorker.Start(ctx)
	}()

	go func() {
		log.Println("order_service starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)                    //Создаём канал с буфером 1 -- всегда нужен буфер размера один. Тип канала os.Signal — специальный тип для сигналов ОС.
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM) //Говорит рантайму Go: "не убивай программу сам, а отправь сигнал в канал quit — я сам разберусь".
	<-quit                                             //Ждем

	cancel()
	wg.Wait() //Это для воркера на отправку в кафку
	kafkaProd.Close()

}
