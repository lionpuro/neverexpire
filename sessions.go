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

func (s *SessionStore) GetUser(r *http.Request) (model.SessionUser, error) {
	sess, err := s.GetSession(r)
	if err != nil {
		return model.SessionUser{}, err
	}
	val := sess.Values["user"]
	user, ok := val.(model.SessionUser)
	if !ok {
		return model.SessionUser{}, fmt.Errorf("invalid session data")
	}
	return user, nil
}

func (s *Server) sessionMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, err := s.Sessions.GetUser(r)
		if err == nil {
			ctx = withUserCtx(r.Context(), user)
		}
		next(w, r.WithContext(ctx))
	}
}

func withUserCtx(ctx context.Context, user model.SessionUser) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func getUserCtx(ctx context.Context) (model.SessionUser, bool) {
	u, ok := ctx.Value(userContextKey).(model.SessionUser)
	return u, ok
}
