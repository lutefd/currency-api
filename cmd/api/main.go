package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/Lutefd/challenge-bravo/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")
	srv := server.NewServer()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	err := srv.Start(ctx)
	if err != nil {
		panic(err)
	}

}
