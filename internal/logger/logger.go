package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Lutefd/challenge-bravo/internal/model"
	"github.com/Lutefd/challenge-bravo/internal/repository"
	"github.com/google/uuid"
)

var (
	InfoLogger          *log.Logger
	ErrorLogger         *log.Logger
	logChan             chan model.Log
	logRepo             repository.LogRepository
	loggerBufferSize    = 1000
	LoggerSleepDuration = 100 * time.Millisecond
)

func init() {
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	logChan = make(chan model.Log, loggerBufferSize)
}

func InitLogger(repo repository.LogRepository) {
	logRepo = repo
	go processLogs()
}

func processLogs() {
	for logEntry := range logChan {
		if err := logRepo.SaveLog(context.Background(), logEntry); err != nil {
			ErrorLogger.Printf("failed to save log: %v", err)
		}
	}
}

func logAsync(level model.LogLevel, message string) {
	logEntry := model.Log{
		ID:        uuid.New(),
		Level:     level,
		Message:   message,
		Timestamp: time.Now(),
		Source:    "application",
	}

	select {
	case logChan <- logEntry:
	default:
		ErrorLogger.Printf("log channel full. Dropping log: %v", logEntry)
	}

	if level == model.LogLevelInfo {
		InfoLogger.Println(message)
	} else {
		ErrorLogger.Println(message)
	}
}

func Info(v ...interface{}) {
	logAsync(model.LogLevelInfo, fmt.Sprint(v...))
}

func Infof(format string, v ...interface{}) {
	logAsync(model.LogLevelInfo, fmt.Sprintf(format, v...))
}

func Error(v ...interface{}) {
	logAsync(model.LogLevelError, fmt.Sprint(v...))
}

func Errorf(format string, v ...interface{}) {
	logAsync(model.LogLevelError, fmt.Sprintf(format, v...))
}

func Shutdown(ctx context.Context) error {
	close(logChan)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if len(logChan) == 0 {
				return logRepo.Close()
			}
			time.Sleep(LoggerSleepDuration)
		}
	}
}
