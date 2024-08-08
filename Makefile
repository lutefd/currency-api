include .env

# Build the application
all: build

build:
	@echo "Building..."
	@go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if [ -x "$(GOPATH)/bin/air" ]; then \
	    "$(GOPATH)/bin/air"; \
		@echo "Watching...";\
	else \
	    read -p "air is not installed. Do you want to install it now? (y/n) " choice; \
	    if [ "$$choice" = "y" ]; then \
			go install github.com/cosmtrek/air@latest; \
	        "$(GOPATH)/bin/air"; \
				@echo "Watching...";\
	    else \
	        echo "You chose not to install air. Exiting..."; \
	        exit 1; \
	    fi; \
	fi

# Docker commands
docker-build:
	docker build -t myapp .

docker-run:
	docker run -p 8080:8080 myapp

# DB commands
migrate-up:
	goose -dir ./sql/schema postgres "$(POSTGRES_CONN)" up

migrate-down:
	goose -dir ./sql/schema postgres "$(POSTGRES_CONN)" down

seed:
	go run ./sql/seed.go

.PHONY: all build run test clean watch docker-build docker-run migrate-up migrate-down seed
