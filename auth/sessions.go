package auth

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/lionpuro/neverexpire/internal/redisstore"
	"github.com/lionpuro/neverexpire/model"
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

	gob.Register(model.User{})

	return &SessionStore{store}, err
}

func (s *SessionStore) GetSession(r *http.Request) (*sessions.Session, error) {
	return s.store.Get(r, "user-session")
}

func (s *SessionStore) GetUser(r *http.Request) (model.User, error) {
	sess, err := s.GetSession(r)
	if err != nil {
		return model.User{}, err
	}
	val := sess.Values["user"]
	user, ok := val.(model.User)
	if !ok {
		return model.User{}, fmt.Errorf("invalid session data")
	}
	return user, nil
}
