.PHONY: all
all: ;

DATABASE_DSN="postgres://postgres:P@ssw0rd@localhost/gophkeeper?sslmode=disable"

.PHONY: up
up:
	@docker-compose up

.PHONY: down
down:
	@docker-compose down

.PHONY: migrate-up
migrate-up:
	migrate -database $(DATABASE_DSN) -path db/migrations up

.PHONY: migrate-down
migrate-down:
	migrate -database $(DATABASE_DSN) -path db/migrations down

.PHONY: clean-data
clean-data:
	sudo rm -rf ./db/data/

.PHONY: lint
lint:
	golangci-lint run --fix

.PHONY: test
test:
	go test -v ./...
