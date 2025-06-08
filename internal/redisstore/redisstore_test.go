package redisstore

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/lionpuro/neverexpire/config"
	"github.com/redis/go-redis/v9"
)

func newTestClient() (*redis.Client, error) {
	conf, err := config.FromEnvFile("../../.env.test")
	if err != nil {
		return nil, fmt.Errorf("failed to load test config: %v", err)
	}
	client := redis.NewClient(&redis.Options{
		Addr: conf.RedisURL,
	})
	return client, nil
}

func TestNew(t *testing.T) {
	client, err := newTestClient()
	if err != nil {
		t.Fatal("failed to create test client", err)
	}

	store, err := NewRedisStore(context.Background(), client)
	if err != nil {
		t.Fatal("failed to create redis store", err)
	}

	req, err := http.NewRequest("GET", "http://www.example.com", nil)
	if err != nil {
		t.Fatal("failed to create request", err)
	}

	session, err := store.New(req, "hello")
	if err != nil {
		t.Fatal("failed to create session", err)
	}
	if session.IsNew == false {
		t.Fatal("session is not new")
	}
}

func TestOptions(t *testing.T) {
	client, err := newTestClient()
	if err != nil {
		t.Fatal("failed to create test client", err)
	}

	store, err := NewRedisStore(context.Background(), client)
	if err != nil {
		t.Fatal("failed to create redis store", err)
	}

	opts := sessions.Options{
		Path:   "/path",
		MaxAge: 99999,
	}
	store.Options(opts)

	req, err := http.NewRequest("GET", "http://www.example.com", nil)
	if err != nil {
		t.Fatal("failed to create request", err)
	}

	session, err := store.New(req, "hello")
	if err != nil {
		t.Fatal("failed to create store", err)
	}
	if session.Options.Path != opts.Path || session.Options.MaxAge != opts.MaxAge {
		t.Fatal("failed to set options")
	}
}

func TestSave(t *testing.T) {
	client, err := newTestClient()
	if err != nil {
		t.Fatal("failed to create test client", err)
	}

	store, err := NewRedisStore(context.Background(), client)
	if err != nil {
		t.Fatal("failed to create redis store", err)
	}

	req, err := http.NewRequest("GET", "http://www.example.com", nil)
	if err != nil {
		t.Fatal("failed to create request", err)
	}
	w := httptest.NewRecorder()

	session, err := store.New(req, "hello")
	if err != nil {
		t.Fatal("failed to create session", err)
	}

	session.Values["key"] = "value"
	err = session.Save(req, w)
	if err != nil {
		t.Fatal("failed to save: ", err)
	}
}

func TestDelete(t *testing.T) {
	client, err := newTestClient()
	if err != nil {
		t.Fatal("failed to create test client", err)
	}

	store, err := NewRedisStore(context.Background(), client)
	if err != nil {
		t.Fatal("failed to create redis store", err)
	}

	req, err := http.NewRequest("GET", "http://www.example.com", nil)
	if err != nil {
		t.Fatal("failed to create request", err)
	}
	w := httptest.NewRecorder()

	session, err := store.New(req, "hello")
	if err != nil {
		t.Fatal("failed to create session", err)
	}

	session.Values["key"] = "value"
	err = session.Save(req, w)
	if err != nil {
		t.Fatal("failed to save session: ", err)
	}

	session.Options.MaxAge = -1
	err = session.Save(req, w)
	if err != nil {
		t.Fatal("failed to delete session: ", err)
	}
}

func TestClose(t *testing.T) {
	client, err := newTestClient()
	if err != nil {
		t.Fatal("failed to create test client", err)
	}

	cmd := client.Ping(context.Background())
	if cmd.Err() != nil {
		t.Fatal("connection is not opened")
	}

	store, err := NewRedisStore(context.Background(), client)
	if err != nil {
		t.Fatal("failed to create redis store", err)
	}

	err = store.Close()
	if err != nil {
		t.Fatal("failed to close")
	}

	cmd = client.Ping(context.Background())
	if cmd.Err() == nil {
		t.Fatal("connection is properly closed")
	}
}
