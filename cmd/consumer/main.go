package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"

	"payment_ms/application"
	infradb "payment_ms/infrastructure/db"
	prisma "payment_ms/infrastructure/db/prisma"
	infrakafka "payment_ms/infrastructure/kafka"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	prismaClient := prisma.NewClient()
	if err := prismaClient.Prisma.Connect(); err != nil {
		log.Fatalf("prisma connect: %v", err)
	}
	defer func() {
		if err := prismaClient.Prisma.Disconnect(); err != nil {
			log.Printf("prisma disconnect: %v", err)
		}
	}()

	inboxRepo := infradb.NewInboxRepository(prismaClient)
	uc := application.NewHandleDeliveryConfirmedUseCase(inboxRepo)

	brokers := strings.Split(env("KAFKA_BROKERS", "localhost:9092"), ",")
	consumer := infrakafka.NewDeliveryConfirmedConsumer(brokers, uc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		cancel()
	}()

	log.Println("delivery.confirmed consumer started")
	if err := consumer.Run(ctx); err != nil {
		log.Fatalf("consumer error: %v", err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
