package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/steebchen/prisma-client-go/cli"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	name := os.Getenv("MIGRATION_NAME")
	if name == "" {
		name = "migrate"
	}

	if err := cli.Run([]string{
		"migrate", "dev",
		"--name", name,
		"--schema", "prisma/schema.prisma",
	}, true); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}
