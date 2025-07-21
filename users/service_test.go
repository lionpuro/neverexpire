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
var currentUser users.User

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
	user, err := testutils.NewTestUser()
	if err != nil {
		t.Errorf("failed to create test data: %v", err)
		return
	}
	err = service.Create(user.ID, user.Email)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	currentUser = user
}

func TestGetUser(t *testing.T) {
	_, err := service.ByID(context.Background(), currentUser.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSaveSettings(t *testing.T) {
	wh := "webhook.example.com"
	th := notifications.ThresholdWeek
	_, err := service.SaveSettings(currentUser.ID, users.SettingsInput{
		WebhookURL:   &wh,
		RemindBefore: &th,
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetSettings(t *testing.T) {
	_, err := service.Settings(context.Background(), currentUser.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDeleteUser(t *testing.T) {
	err := service.Delete(currentUser.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
