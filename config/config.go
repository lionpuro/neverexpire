package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	OAuthGoogleClientID,
	OAuthGoogleClientSecret,
	OAuthGoogleCallbackURL,
	RedisURL,
	PostgresUser,
	PostgresPassword,
	PostgresHost,
	PostgresDB,
	PostgresPort string
}

func FromEnv() *Config {
	conf := &Config{
		OAuthGoogleClientID:     os.Getenv("OAUTH_GOOGLE_CLIENT_ID"),
		OAuthGoogleClientSecret: os.Getenv("OAUTH_GOOGLE_CLIENT_SECRET"),
		OAuthGoogleCallbackURL:  os.Getenv("OAUTH_GOOGLE_CALLBACK_URL"),
		RedisURL:                os.Getenv("REDIS_URL"),
		PostgresUser:            os.Getenv("POSTGRES_USER"),
		PostgresPassword:        os.Getenv("POSTGRES_PASSWORD"),
		PostgresHost:            os.Getenv("POSTGRES_HOST"),
		PostgresDB:              os.Getenv("POSTGRES_DB"),
		PostgresPort:            os.Getenv("POSTGRES_HOST_PORT"),
	}
	return conf
}

func FromEnvFile(file string) (*Config, error) {
	if err := godotenv.Load(file); err != nil {
		return nil, err
	}
	conf := FromEnv()
	return conf, nil
}
