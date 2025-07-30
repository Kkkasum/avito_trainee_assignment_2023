LOCAL_BIN := $(CURDIR)/bin

MIGRATIONS_DIR := $(CURDIR)/migrations

MODULE := $(shell go list -m)

DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_ADDRESS ?= localhost
DB_PORT ?= 5436
DB_NAME ?= avito_2023

.bin-deps: export GOBIN := $(LOCAL_BIN)
.bin-deps:
	$(info Installing binary dependencies...)

	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/matryer/moq@latest

.migrate:
	docker run --rm -v ./migrations:/migrations --network host migrate/migrate -path=/migrations \
	-database postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_ADDRESS):$(DB_PORT)/$(DB_NAME)?sslmode=disable up 1

.gen-swagger-docs:
	./bin/swag init -g cmd/server/main.go

.tidy:
	GOBIN=$(LOCAL_BIN) go mod tidy

.swagger: .bin-deps .gen-swagger-docs .tidy

.build:
	go build -a -o app $(MODULE)/cmd/server

.build-docker:
	docker build -f cmd/server/Dockerfile -t avito/trainee_assignment_2023 .