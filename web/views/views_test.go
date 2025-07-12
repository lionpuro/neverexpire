package views_test

import (
	"bytes"
	"testing"

	"github.com/lionpuro/neverexpire/hosts"
	"github.com/lionpuro/neverexpire/keys"
	"github.com/lionpuro/neverexpire/user"
	"github.com/lionpuro/neverexpire/web/views"
)

func TestRender(t *testing.T) {
	testUser := &user.User{
		Email: "tester@neverexpire.xyz",
	}
	testHosts := []hosts.Host{
		{
			ID:          1,
			HostName:    "neverexpire.xyz",
			Certificate: hosts.CertificateInfo{},
		},
	}
	// Home
	t.Run("home (logged out)", func(t *testing.T) {
		buf := bytes.Buffer{}
		err := views.Home(&buf, views.LayoutData{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("home (logged in)", func(t *testing.T) {
		buf := bytes.Buffer{}
		err := views.Home(&buf, views.LayoutData{User: testUser})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	// Hosts
	t.Run("hosts", func(t *testing.T) {
		buf := bytes.Buffer{}
		err := views.Hosts(&buf, views.LayoutData{User: testUser}, testHosts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	// Host
	t.Run("host", func(t *testing.T) {
		buf := bytes.Buffer{}
		err := views.Host(&buf, views.LayoutData{User: testUser}, testHosts[0])
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	// NewHost
	t.Run("new host", func(t *testing.T) {
		buf := bytes.Buffer{}
		err := views.NewHosts(&buf, views.LayoutData{User: testUser}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	// Settings
	t.Run("settings", func(t *testing.T) {
		buf := bytes.Buffer{}
		err := views.Settings(&buf, views.LayoutData{User: testUser}, user.Settings{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	// API
	t.Run("api", func(t *testing.T) {
		err := views.API(&bytes.Buffer{}, views.LayoutData{User: testUser}, []keys.AccessKey{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	// Login
	t.Run("login", func(t *testing.T) {
		buf := bytes.Buffer{}
		err := views.Login(&buf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
