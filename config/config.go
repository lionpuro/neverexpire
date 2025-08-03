package config

import (
	"fmt"
	"os"
)

type Config struct {
	OAuthGoogleClientID,
	OAuthGoogleClientSecret,
	OAuthGoogleCallbackURL,
	RedisURL,
	RedisPassword,
	PostgresURL string
}

func FromEnv() *Config {
	pgurl := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_HOST_PORT"),
		os.Getenv("POSTGRES_DB"),
	)
	rdurl := fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))
	conf := &Config{
		OAuthGoogleClientID:     os.Getenv("OAUTH_GOOGLE_CLIENT_ID"),
		OAuthGoogleClientSecret: os.Getenv("OAUTH_GOOGLE_CLIENT_SECRET"),
		OAuthGoogleCallbackURL:  os.Getenv("OAUTH_GOOGLE_CALLBACK_URL"),
		RedisURL:                rdurl,
		RedisPassword:           os.Getenv("REDIS_PASSWORD"),
		PostgresURL:             pgurl,
	}
	return conf
}
