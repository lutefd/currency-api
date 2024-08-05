package server

import (
	"os"
	"strconv"
)

type Config struct {
	PostgresConn string
	RedisAddr    string
	RedisPass    string
	ServerPort   uint16
}

func LoadConfig() Config {
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisAddr := os.Getenv("REDIS_ADDR")
	serverPort := os.Getenv("SERVER_PORT")
	postgresConn := os.Getenv("POSTGRES_CONN")
	parsedServerPort, err := strconv.ParseUint(serverPort, 10, 16)
	if err != nil {
		panic(err)
	}

	return Config{
		PostgresConn: postgresConn,
		RedisAddr:    redisAddr,
		RedisPass:    redisPassword,
		ServerPort:   uint16(parsedServerPort),
	}
}
