package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/lionpuro/trackcert/model"
	"github.com/lionpuro/trackcert/redisstore"
	"github.com/redis/go-redis/v9"
)

type SessionStore struct {
	store *redisstore.RedisStore
}

func newSessionStore() (*SessionStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: "",
		DB:       0,
	})
	store, err := redisstore.NewRedisStore(context.Background(), client)
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   true,
		SameSite: 4,
	})

	gob.Register(model.SessionUser{})

	return &SessionStore{store}, err
}

func (s *SessionStore) GetSession(r *http.Request) (*sessions.Session, error) {
	return s.store.Get(r, "user-session")
}

func (s *SessionStore) GetUser(r *http.Request) (*model.SessionUser, error) {
	sess, err := s.GetSession(r)
	if err != nil {
		return nil, err
	}
	val := sess.Values["user"]
	user, ok := val.(model.SessionUser)
	if !ok {
		return nil, fmt.Errorf("invalid session data")
	}
	return &user, nil
}
