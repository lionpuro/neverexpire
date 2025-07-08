package auth

import (
	"context"
	"encoding/gob"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/lionpuro/neverexpire/auth/redisstore"
	"github.com/lionpuro/neverexpire/user"
	"github.com/redis/go-redis/v9"
)

type Session struct {
	session *sessions.Session
}

func (s *Session) User() *user.User {
	user, ok := s.session.Values["user"].(user.User)
	if !ok {
		return nil
	}
	return &user
}

func (s *Session) SetUser(user user.User) {
	s.session.Values["user"] = user
}

func (s *Session) State() (string, bool) {
	state, ok := s.session.Values["state"].(string)
	if !ok {
		return "", false
	}
	return state, true
}

func (s *Session) SetState(value string) {
	s.session.Values["state"] = value
}

func (s *Session) Save(w http.ResponseWriter, r *http.Request) error {
	return s.session.Save(r, w)
}

func (s *Session) Delete(w http.ResponseWriter, r *http.Request) error {
	s.session.Options.MaxAge = -1
	return s.session.Save(r, w)
}

type SessionStore struct {
	store *redisstore.RedisStore
}

func newSessionStore(addr string) (*SessionStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
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

	gob.Register(user.User{})

	return &SessionStore{store}, err
}

func (s *SessionStore) GetSession(r *http.Request) (*Session, error) {
	sess, err := s.store.Get(r, "user-session")
	if err != nil {
		return nil, err
	}
	return &Session{session: sess}, nil
}
