include .env

DATABASE=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_HOST_PORT}/${POSTGRES_DB}?sslmode=disable

.PHONY: all assets app service run-app run-service fmt lint create-migration migrate-up migrate-down

all: assets app service

assets:
	@npm run build
	@npx @tailwindcss/cli -i ./assets/src/global.css -o ./assets/public/css/global.css --minify

app:
	@go build -o tmp/app ./cmd/app

service:
	@go build -o tmp/service ./cmd/service

run-app:
	@./tmp/app

run-service:
	@./tmp/service

dev:
	@air -c .air.toml

fmt:
	@gofmt -l -s -w .
	@npx prettier . --write

lint:
	@docker run -t --rm -v ${PWD}:/app -w /app golangci/golangci-lint:v2.1.6 golangci-lint run

create-migration:
	@read -p "Enter the sequence name: " SEQ; \
		docker run -u 1000:1000 --rm -v ./migrations:/migrations migrate/migrate \
			create -ext sql -dir /migrations -seq $${SEQ}

migrate-up:
	@docker run --rm -v ./migrations:/migrations --network host migrate/migrate \
		-path=/migrations -database "${DATABASE}" up

migrate-down:
	@read -p "Number of migrations you want to rollback (default: 1): " NUM; NUM=$${NUM:-1}; \
		docker run --rm -it -v ./migrations:/migrations --network host migrate/migrate \
			-path=/migrations -database "${DATABASE}" down $${NUM}
