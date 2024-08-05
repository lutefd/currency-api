package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Server struct {
	port   int
	router http.Handler
	config Config
}

func NewServer(config Config) *Server {
	strPort := os.Getenv("SERVER_PORT")
	if strPort == "" {
		fmt.Println("port environment variable was not setted")
	}
	server := &Server{
		port:   int(config.ServerPort),
		config: config,
	}
	server.registerRoutes()
	return server
}

func (s *Server) Start(ctx context.Context) error {
	fmt.Println("Starting server on port 8080")
	ch := make(chan error, 1)
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	var err error
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		close(ch)
	}()
	select {
	case err = <-ch:
		return err
	case <-ctx.Done():
		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		return server.Shutdown(ctx)
	}
}
