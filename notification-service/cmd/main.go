package main

import (
	"context"
	"log"
	"notification-service/internal/consumer"
	"notification-service/internal/notifier"
	"notification-service/internal/repository"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func main() {

	//
	db, err := sqlx.Connect("pgx", "postgres://postgres:postgres@localhost:5433/notifications_db")
	if err != nil {
		log.Fatal("failed to connect to db:", err)
	}
	defer db.Close()
	log.Println("connected to database")

	//Соберем все слои ( как пирог: сначала репозиторий от db, затем notifier от репозитория)
	repo := repository.NewNotificationRepository(db)
	n := notifier.NewNotifier(repo)

	c := consumer.NewKafkaConsumer(
		"localhost:9092",
		"orders",
		"notification_service",
		n,
	)

	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done() // сигнализируем что горутина завершилась
		c.Start(ctx)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	cancel()  // говорим горутине остановиться
	wg.Wait() // ждём пока она реально остановится
	log.Println("shutdown complete")
}
