package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	address := flag.String("address", ":8080", "HTTP listen address")
	postgresURL := flag.String("postgres-url", "postgres://sen:sen@127.0.0.1:5432/sen?sslmode=disable", "PostgreSQL connection string")
	redisAddress := flag.String("redis-address", "127.0.0.1:6379", "Redis address")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbPool, err := pgxpool.New(ctx, *postgresURL)
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: *redisAddress})
	defer func() {
		if closeErr := redisClient.Close(); closeErr != nil {
			log.Printf("close redis client: %v", closeErr)
		}
	}()

	app := &application{
		service: visitService{
			store: visitStore{db: dbPool},
			cache: visitCache{client: redisClient},
		},
	}

	server := &http.Server{
		Addr:              *address,
		Handler:           app.routes(),
		ReadHeaderTimeout: 2 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Printf("HTTP server listening on %s", *address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

type application struct {
	service visitService
}
