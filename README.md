# NeverExpire

Automated TLS certificate status monitoring service for the people who manage their certs manually.

[App screenshot](/assets/static/images/hero.webp)

## Features

- Regular scanning of tracked hosts for certificate expiry and status
- Configurable notifications via webhooks
- API for managing tracked hosts

## Development

Start the development containers using the dev helper script:

```sh
./dev.sh start
```

or with docker compose directly:

```sh
docker compose -f compose.dev.yaml up
```

This starts the web app and the background process in the workspace container as well as all the
dependencies. The app is reloaded during development using [wgo](https://github.com/bokwoon95/wgo).
