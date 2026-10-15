BIN_DIR := "./bin"
CALENDAR_BIN := "$(BIN_DIR)/calendar"
SCHEDULER_BIN := "$(BIN_DIR)/calendar_scheduler"
SENDER_BIN := "$(BIN_DIR)/calendar_sender"

DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git rev-parse --short HEAD)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

#Postrges
POSTGRES_USER ?= postgres
POSTGRES_PASSWORD ?= postgres
POSTGRES_DB ?= postgres
POSTGRES_PORT ?= 5435
POSTGRES_CONTAINER := postgres-calendar

#Rabbit
RABBIT_USER ?= rabbit
RABBIT_PASSWORD ?= password
RABBIT_PORT ?= 5672
RABBIT_PORT_UI ?= 15672
RABBIT_CONTAINER := rabbitmq

run-postgres:
	docker run -d --name $(POSTGRES_CONTAINER) \
	-e POSTGRES_USER=$(POSTGRES_USER) \
	-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
	-e POSTGRES_DB=$(POSTGRES_DB) \
	-p $(POSTGRES_PORT):5432 \
	-v postgres-data:/var/lib/postgresql/data \
	postgres:latest

migrate:
	goose -dir ./migrations postgres "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" up

migrate-down:
	goose -dir ./migrations postgres "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" down

migrate-status:
	goose -dir ./migrations postgres "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" status

run-rabbit:
	docker run -d --name $(RABBIT_CONTAINER) \
	-p $(RABBIT_PORT):5672 \
	-p $(RABBIT_PORT_UI):15672 \
	-e RABBITMQ_DEFAULT_USER=$(RABBIT_USER) \
	-e RABBITMQ_DEFAULT_PASS=$(RABBIT_PASSWORD) \
	rabbitmq:management

build:
	go build -v -o $(CALENDAR_BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar
	go build -v -o $(SCHEDULER_BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar_scheduler
	go build -v -o $(SENDER_BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar_sender

run: build
	$(CALENDAR_BIN) -config ./configs/config.toml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(CALENDAR_BIN) version

test:
	go test -race ./internal/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v2.13.2

lint: install-lint-deps
	golangci-lint run ./...

.PHONY: build run build-img run-img version test lint generate \
	run-postgres run-rabbit migrate migrate-down migrate-status \
	up down integration-tests

generate:
	go generate ./...

COMPOSE_FILE := ./deployments/docker-compose.yaml
INTEGRATION_COMPOSE_FILE := ./deployments/docker-compose.integration.yaml

up:
	docker compose -f $(COMPOSE_FILE) up -d

down:
	docker compose -f $(COMPOSE_FILE) down

integration-tests:
	docker compose -f $(INTEGRATION_COMPOSE_FILE) up -d --build
	@while [ "$$(docker inspect -f '{{.State.Status}}' $$(docker compose -f $(INTEGRATION_COMPOSE_FILE) ps -q integration-tests))" != "exited" ]; do \
		sleep 1; \
	done
	@code=$$(docker inspect -f '{{.State.ExitCode}}' $$(docker compose -f $(INTEGRATION_COMPOSE_FILE) ps -q integration-tests)); \
	docker compose -f $(INTEGRATION_COMPOSE_FILE) down -v; \
	exit $$code