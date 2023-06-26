.PHONY: all run
all: tests lintcheck build

run: dbrun devrun

devrun:
	export `grep -v '#' .env | xargs` && DB_HOST=localhost && go run cmd/main.go

dbrun:
	docker-compose up -d db

migration-up:
	docker-compose up -d db migrate-up

migration-down:
	docker-compose -f migrate-down.yml up -d

build:
	docker-compose up -d --build

up:
	docker-compose up -d

down:
	docker-compose down

lintcheck:
	golangci-lint run

test:
	go test ./...

coverage:
	go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html	&& xdg-open ./coverage.html