ARG GO_VERSION=1.24.4

# Assets
FROM node:22 AS assets
WORKDIR /app
COPY . .
RUN npm install
RUN npm run build

# Build
FROM golang:${GO_VERSION} AS build
WORKDIR /src
COPY . .
RUN go mod download && go mod verify
RUN go build -v -o /out/app ./cmd/app
RUN go build -v -o /out/worker ./cmd/worker

# Run
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates
WORKDIR /app
COPY --from=build /out/app .
COPY --from=build /out/worker .
COPY --from=assets /app/assets/public ./assets/public
