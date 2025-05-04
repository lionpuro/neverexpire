.PHONY: setup build run fmt

setup:
	mkdir -p .tools
	curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64
	chmod +x tailwindcss-linux-x64
	mv tailwindcss-linux-x64 .tools/tailwindcss

build:
	@./.tools/tailwindcss -i ./global.css -o ./static/css/global.css --minify
	@go build -o tmp/run .

run: build
	@./tmp/run

fmt:
	@gofmt -l -s -w .
