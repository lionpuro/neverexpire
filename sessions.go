package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/lionpuro/trackcerts/model"
	"github.com/lionpuro/trackcerts/redisstore"
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

func withUserCtx(ctx context.Context, user model.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func getUserCtx(ctx context.Context) (model.User, bool) {
	u, ok := ctx.Value(userContextKey).(model.User)
	return u, ok
}
