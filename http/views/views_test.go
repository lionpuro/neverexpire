package views_test

import (
	"bytes"
	"testing"

	"github.com/lionpuro/neverexpire/http/views"
	"github.com/lionpuro/neverexpire/model"
)

func TestRender(t *testing.T) {
	testUser := &model.User{
		Email: "tester@neverexpire.xyz",
	}
	testDomains := []model.Domain{
		{
			ID:          1,
			DomainName:  "neverexpire.xyz",
			Certificate: model.CertificateInfo{},
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
	// Domains
	t.Run("domains", func(t *testing.T) {
		buf := bytes.Buffer{}
		err := views.Domains(&buf, views.LayoutData{User: testUser}, testDomains)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	// Domain
	t.Run("domain", func(t *testing.T) {
		buf := bytes.Buffer{}
		err := views.Domain(&buf, views.LayoutData{User: testUser}, testDomains[0])
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	// NewDomain
	t.Run("new domain", func(t *testing.T) {
		buf := bytes.Buffer{}
		err := views.NewDomains(&buf, views.LayoutData{User: testUser}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	// Settings
	t.Run("settings", func(t *testing.T) {
		buf := bytes.Buffer{}
		err := views.Settings(&buf, views.LayoutData{User: testUser}, model.Settings{})
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
