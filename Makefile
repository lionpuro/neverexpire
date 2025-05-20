include .env

DATABASE=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_HOST_PORT}/${POSTGRES_DB}?sslmode=disable

.PHONY: setup build run fmt create-migration migrate-up migrate-down

setup:
	mkdir -p .tools
	curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64
	chmod +x tailwindcss-linux-x64
	mv tailwindcss-linux-x64 .tools/tailwindcss

build:
	@npm run build
	@./.tools/tailwindcss -i ./global.css -o ./assets/public/css/global.css --minify
	@go build -o tmp/run .

run: build
	@./tmp/run

dev:
	@air -c .air.toml

fmt:
	@gofmt -l -s -w .
	@npx prettier . --write

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
