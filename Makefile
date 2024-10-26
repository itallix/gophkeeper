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

.PHONY: build-server
build-server:
	go build -o ./bin/server ./cmd/server

.PHONY: run-server
run-server:
	./bin/server

.PHONY: build-client
build-client:
	go build -o ./bin/client ./cmd/client

.PHONY: run-client
run-client:
	./bin/client

.PHONY: generate
generate:
	protoc --go_out=pkg/generated --go_opt=paths=source_relative \
		--go-grpc_out=pkg/generated --go-grpc_opt=paths=source_relative \
		api/proto/v1/*.proto
