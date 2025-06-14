include .env

DATABASE=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_HOST_PORT}/${POSTGRES_DB}?sslmode=disable

.PHONY: all assets app service run-app run-service dev-up fmt lint test docker-deploy create-migration migrate-up migrate-down

all: assets app service

assets:
	@npm run build

app:
	@go build -o tmp/app ./cmd/app

service:
	@go build -o tmp/service ./cmd/service

run-app:
	@./tmp/app

run-service:
	@./tmp/service

dev-up:
	@docker compose -f compose.dev.yaml up

watch:
	@wgo -debounce 100ms -xdir assets/public clear :: npm run build:tw :: go run ./cmd/app :: wgo go run ./cmd/service

watch-app:
	@wgo -debounce 100ms -xdir assets/public clear :: npm run build:tw :: go build -o tmp/app ./cmd/app :: ./tmp/app

fmt:
	@gofmt -l -s -w .
	@npx prettier . --write

lint:
	@docker run -t --rm -v ${PWD}:/app -w /app golangci/golangci-lint:v2.1.6 golangci-lint run

test:
	@go test -v ./...

docker-deploy:
	DOCKER_CONTEXT=neverexpire docker compose up --build -d

create-migration:
	@read -p "Enter the sequence name: " SEQ; \
		docker run -u 1000:1000 --rm -v ./db/migrations:/migrations migrate/migrate \
			create -ext sql -dir /migrations -seq $${SEQ}

migrate-up:
	@docker run --rm -v ./db/migrations:/migrations --network host migrate/migrate \
		-path=/migrations -database "${DATABASE}" up

migrate-down:
	@read -p "Number of migrations you want to rollback (default: 1): " NUM; NUM=$${NUM:-1}; \
		docker run --rm -it -v ./db/migrations:/migrations --network host migrate/migrate \
			-path=/migrations -database "${DATABASE}" down $${NUM}

migrate-force:
	@read -p "Enter the version to force: " VERSION; \
	docker run -u 1000:1000 --rm -it -v ./db/migrations:/migrations --network host migrate/migrate \
		-path=/migrations -database "${DATABASE}" force $${VERSION}
