package redisstore

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/gob"
	"errors"
	"github.com/gorilla/sessions"
	"github.com/redis/go-redis/v9"
	"io"
	"net/http"
	"strings"
	"time"
)

type RedisStore struct {
	client     redis.UniversalClient
	options    sessions.Options
	keyPrefix  string
	keyGen     KeyGenFunc
	serializer SessionSerializer
}

type KeyGenFunc func() (string, error)

func NewRedisStore(ctx context.Context, client redis.UniversalClient) (*RedisStore, error) {
	rs := &RedisStore{
		options: sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 30,
			HttpOnly: true,
			Secure:   true,
			SameSite: 4,
		},
		client:     client,
		keyPrefix:  "session:",
		keyGen:     generateRandomKey,
		serializer: GobSerializer{},
	}

	return rs, rs.client.Ping(ctx).Err()
}

func (s *RedisStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

func (s *RedisStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	opts := s.options
	session.Options = &opts
	session.IsNew = true

	c, err := r.Cookie(name)
	if err != nil {
		return session, nil
	}
	session.ID = c.Value

	err = s.load(r.Context(), session)
	switch err {
	case nil:
		session.IsNew = false
	case redis.Nil:
		err = nil // no data stored
	}
	return session, err
}

// If the Options.MaxAge of the session is <= 0 then the session file will be
// deleted from the store.
func (s *RedisStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	if session.Options.MaxAge <= 0 {
		if err := s.delete(r.Context(), session); err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
		return nil
	}

	if session.ID == "" {
		id, err := s.keyGen()
		if err != nil {
			return errors.New("redisstore: failed to generate session id")
		}
		session.ID = id
	}
	if err := s.save(r.Context(), session); err != nil {
		return err
	}

	http.SetCookie(w, sessions.NewCookie(session.Name(), session.ID, session.Options))
	return nil
}

func (s *RedisStore) Options(opts sessions.Options) {
	s.options = opts
}

func (s *RedisStore) KeyPrefix(keyPrefix string) {
	s.keyPrefix = keyPrefix
}

func (s *RedisStore) KeyGen(f KeyGenFunc) {
	s.keyGen = f
}

func (s *RedisStore) Serializer(ss SessionSerializer) {
	s.serializer = ss
}

func (s *RedisStore) Close() error {
	return s.client.Close()
}

func (s *RedisStore) save(ctx context.Context, session *sessions.Session) error {
	b, err := s.serializer.Serialize(session)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, s.keyPrefix+session.ID, b, time.Duration(session.Options.MaxAge)*time.Second).Err()
}

func (s *RedisStore) load(ctx context.Context, session *sessions.Session) error {
	cmd := s.client.Get(ctx, s.keyPrefix+session.ID)
	if cmd.Err() != nil {
		return cmd.Err()
	}

	b, err := cmd.Bytes()
	if err != nil {
		return err
	}

	return s.serializer.Deserialize(b, session)
}

func (s *RedisStore) delete(ctx context.Context, session *sessions.Session) error {
	return s.client.Del(ctx, s.keyPrefix+session.ID).Err()
}

type SessionSerializer interface {
	Serialize(s *sessions.Session) ([]byte, error)
	Deserialize(b []byte, s *sessions.Session) error
}

type GobSerializer struct{}

func (gs GobSerializer) Serialize(s *sessions.Session) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(s.Values)
	if err == nil {
		return buf.Bytes(), nil
	}
	return nil, err
}

func (gs GobSerializer) Deserialize(d []byte, s *sessions.Session) error {
	dec := gob.NewDecoder(bytes.NewBuffer(d))
	return dec.Decode(&s.Values)
}

func generateRandomKey() (string, error) {
	k := make([]byte, 64)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return "", err
	}
	return strings.TrimRight(base32.StdEncoding.EncodeToString(k), "="), nil
}
