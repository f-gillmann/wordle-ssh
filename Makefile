APP_NAME	:= wordle-ssh
CMD_DIR		:= ./cmd
BIN_DIR 	:= ./bin

# Detect OS
ifeq ($(OS),Windows_NT)
	BINARY := $(APP_NAME).exe
	MKDIR_CMD := powershell -Command "if (!(Test-Path $(BIN_DIR))) { New-Item -ItemType Directory -Path $(BIN_DIR) | Out-Null }"
	RM_CMD := powershell -Command "if (Test-Path $(BIN_DIR)) { Remove-Item -Recurse -Force $(BIN_DIR) }"
	RUN_BINARY := $(BIN_DIR)\$(BINARY)
else
	BINARY := $(APP_NAME)
	MKDIR_CMD := mkdir -p $(BIN_DIR)
	RM_CMD := rm -rf $(BIN_DIR)
	RUN_BINARY := $(BIN_DIR)/$(BINARY)
endif

.PHONY: all build run clean test fmt vet tidy docker-build docker-run docker-stop docker-clean docker-up docker-down

all: build

build:
	@$(MKDIR_CMD)
	go build -o $(BIN_DIR)/$(BINARY) $(CMD_DIR)

run: build
	$(RUN_BINARY)

clean:
	@$(RM_CMD)

tidy:
	go mod tidy

fmt:
	gofmt -w .

vet:
	go vet ./...

test:
	go test ./...

docker-build:
	docker build -t wordle-ssh:latest .

docker-run:
	docker run -d -p 23234:23234 --name wordle-ssh wordle-ssh:latest

docker-stop:
	docker stop wordle-ssh || true
	docker rm wordle-ssh || true

docker-clean: docker-stop
	docker rmi wordle-ssh:latest || true

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down
