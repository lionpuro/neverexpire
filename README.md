# NeverExpire

Automated TLS certificate status monitoring service for the people who manage their certs manually.

![App screenshot](/assets/static/images/screenshot.webp)

## Features

- Regular scanning of tracked hosts for certificate expiry and status
- Configurable notifications via webhooks
- API for managing tracked hosts

## Development

1. Install [Docker](https://docs.docker.com/get-started/)
2. Clone this repo
3. Create a `.env` file in the root of the project using `.env.example` as a template.
4. Start the containers by running `docker compose -f compose.dev.yaml up`
5. Install npm dependencies by running `docker compose -f compose.dev.yaml exec workspace npm install`
6. neverexpire should be running on `localhost:3000`
