package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/Lutefd/challenge-bravo/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")
	config, err := server.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	srv := server.NewServer(config)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	err = srv.Start(ctx)
	if err != nil {
		panic(err)
	}

}
