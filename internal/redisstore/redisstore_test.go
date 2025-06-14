package redisstore_test

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/lionpuro/neverexpire/internal/redisstore"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var client *redis.Client
var store *redisstore.RedisStore

func TestMain(m *testing.M) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	container, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			log.Printf("failed to terminate container: %v", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %v", err)
		return
	}
	endpoint, err := container.Endpoint(context.Background(), "")
	if err != nil {
		log.Printf("failed to get redis client endpoint: %v", err)
		return
	}
	client = redis.NewClient(&redis.Options{
		Addr: endpoint,
	})
	store, err = redisstore.NewRedisStore(context.Background(), client)
	if err != nil {
		log.Printf("failed to create redis store: %v", err)
		return
	}
	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
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
	cmd := client.Ping(context.Background())
	if cmd.Err() != nil {
		t.Fatal("connection is not opened")
	}

	err := store.Close()
	if err != nil {
		t.Fatal("failed to close")
	}

	cmd = client.Ping(context.Background())
	if cmd.Err() == nil {
		t.Fatal("connection is properly closed")
	}
}
