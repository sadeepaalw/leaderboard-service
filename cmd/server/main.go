package main

import (
	"context"
	"leaderboard-service/internal/api"
	"leaderboard-service/internal/db"
	"leaderboard-service/internal/repository"
	"leaderboard-service/internal/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	database := db.Open()
	defer db.Close(database)

	// Set connection pool settings
	database.SetMaxOpenConns(20)
	database.SetMaxIdleConns(10)
	database.SetConnMaxLifetime(30 * time.Second)

	repo := repository.NewRepository(database)
	svc := service.NewService(repo)
	handler := api.NewHandler(svc)

	router := api.NewRouter(handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc.StartMatchmakingWorker(ctx)

	// Graceful shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Println("Shutting down...")
		cancel()
	}()

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
