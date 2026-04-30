package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.temporal.io/sdk/client"

	"payment_ms/application"
	infradb "payment_ms/infrastructure/db"
	prisma "payment_ms/infrastructure/db/prisma"
	kafkaconsumer "payment_ms/infrastructure/kafka"
	infratemporal "payment_ms/infrastructure/temporal"
	httphandler "payment_ms/interfaces/http"
)

func main() {
	// ── Config ────────────────────────────────────────────────────────────────
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	// ── Prisma ────────────────────────────────────────────────────────────────
	prismaClient := prisma.NewClient()
	if err := prismaClient.Prisma.Connect(); err != nil {
		log.Fatalf("prisma connect: %v", err)
	}
	defer func() {
		if err := prismaClient.Prisma.Disconnect(); err != nil {
			log.Printf("prisma disconnect: %v", err)
		}
	}()

	// ── Temporal ──────────────────────────────────────────────────────────────
	temporalClient, err := client.Dial(client.Options{
		HostPort:  env("TEMPORAL_HOST", "localhost:7233"),
		Namespace: env("TEMPORAL_NAMESPACE", "default"),
	})
	if err != nil {
		log.Fatalf("temporal dial: %v", err)
	}
	defer temporalClient.Close()

	// ── Infrastructure ────────────────────────────────────────────────────────
	repo := infradb.NewPaymentRepository(prismaClient)
	inboxRepo := infradb.NewInboxRepository(prismaClient)
	signaler := infratemporal.NewWorkflowSignaler(temporalClient)

	// ── Use cases ─────────────────────────────────────────────────────────────
	initiateUC := application.NewInitiatePaymentUseCase(repo)
	webhookUC := application.NewHandleWebhookUseCase(repo, signaler)
	refundUC := application.NewRequestRefundUseCase(repo)
	handleDeliveryUC := application.NewHandleDeliveryConfirmedUseCase(inboxRepo)

	// Kafka consumer─────────────────────────────────────────────────────────────
	brokers := []string{env("KAFKA_BROKERS", "localhost:9094")}
	dlqProducer := kafkaconsumer.NewKafkaDLQProducer(brokers)
	defer dlqProducer.Close()
	consumer := kafkaconsumer.NewDeliveryConfirmedConsumer(brokers, handleDeliveryUC, dlqProducer, 3)
	go consumer.Run(context.Background())

	// ── HTTP server ───────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         env("HTTP_ADDR", ":8080"),
		Handler:      httphandler.NewRouter(initiateUC, webhookUC, refundUC),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server: %v", err)
		}
	}()

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("http shutdown: %v", err)
	}
	log.Println("server stopped")
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
