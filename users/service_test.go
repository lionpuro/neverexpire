package users_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/lionpuro/neverexpire/notifications"
	"github.com/lionpuro/neverexpire/testutils"
	"github.com/lionpuro/neverexpire/users"
)

var service *users.Service

func TestMain(m *testing.M) {
	conn, cleanup, err := testutils.NewPostgresConn()
	defer func() {
		if err := cleanup(); err != nil {
			log.Printf("error calling cleanup function: %v", err)
		}
	}()
	if err != nil {
		log.Printf("init postgres: %v", err)
		return
	}
	service = users.NewService(users.NewRepository(conn))
	os.Exit(m.Run())
}

func TestCreateUser(t *testing.T) {
	err := service.Create("test-id", "tester@neverexpire.xyz")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetUser(t *testing.T) {
	_, err := service.ByID(context.Background(), "test-id")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSaveSettings(t *testing.T) {
	wh := "webhook.example.com"
	th := notifications.ThresholdWeek
	_, err := service.SaveSettings("test-id", users.SettingsInput{
		WebhookURL:   &wh,
		RemindBefore: &th,
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetSettings(t *testing.T) {
	_, err := service.Settings(context.Background(), "test-id")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDeleteUser(t *testing.T) {
	err := service.Delete("test-id")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
