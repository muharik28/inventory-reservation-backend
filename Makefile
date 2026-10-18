rest:
	go run main.go rest

sch:
	go run main.go scheduler

help:
	go run main.go help

help-rest:
	go run main.go help rest

help-sch:
	go run main.go help scheduler

help-flag:
	go run main.go rest --help

help-completion:
	go run main.go completion --help

app-build:
	go build -o inventory.exe main.go

test-all:
	go test -race -v ./...

test-unit:
	go test -v ./test/...

test-full:
	go test -v ./...

test-integration:
	go test -race -v ./test -run TestDomainLogic

test-specific:
	go test -v ./test -run TestDomainLogic

test-concurrency:
	go test -race -v ./test -run TestConcurrentReservationsDoNotOversell

test-cache:
	go test -count=1 -race -v ./test/...

test-all-cache:
	go test -count=1 -race -v ./...

swag:
	swag init -g cmd/rest/rest.go -o docs/api

compose:
	docker compose -f deployments/docker-compose.yml up -d --build

compose-db:
	docker compose -f deployments/docker-compose.yml up -d postgres

compose-db-restart:
	docker compose -f deployments/docker-compose.yml restart postgres

compose-ps:
	docker compose -f deployments/docker-compose.yml ps

compose-logs:
	docker compose -f deployments/docker-compose.yml logs -f

compose-log-app:
	docker compose -f deployments/docker-compose.yml logs -f app

compose-log-sch:
	docker compose -f deployments/docker-compose.yml logs -f scheduler

compose-down:
	docker compose -f deployments/docker-compose.yml down

compose-down-volume:
	docker compose -f deployments/docker-compose.yml down -v

compose-cache-app:
	docker build --no-cache --progress=plain -t app .

compose-cache-sch:
	docker build --no-cache --progress=plain -t scheduler .

.PHONY: rest sch help help-rest help-sch help-flag help-completion app-build test-all test-unit test-full test-integration test-specific test-concurrency test-cache test-all-cache swag compose compose-db compose-db-restart compose-ps compose-logs compose-log-app compose-log-sch compose-down compose-down-volume compose-cache-app compose-cache-sch