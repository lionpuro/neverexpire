include .env

DATABASE=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_HOST_PORT}/${POSTGRES_DB}?sslmode=disable

.PHONY: lint test docker-deploy create-migration migrate-up migrate-down migrate-force

lint:
	@docker run -t --rm -v ${PWD}:/app -w /app golangci/golangci-lint:v2.2.2 golangci-lint run

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
