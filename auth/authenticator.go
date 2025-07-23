package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/lionpuro/neverexpire/config"
	"golang.org/x/oauth2"
)

type Authenticator struct {
	GoogleClient *Client
	sessions     *SessionStore
}

func NewAuthenticator(conf *config.Config) (*Authenticator, error) {
	googleProvider, err := oidc.NewProvider(context.Background(), "https://accounts.google.com")
	if err != nil {
		return nil, fmt.Errorf("new google provider: %v", err)
	}
	googleClient, err := newAuthClient(googleProvider, &oauth2.Config{
		ClientID:     conf.OAuthGoogleClientID,
		ClientSecret: conf.OAuthGoogleClientSecret,
		RedirectURL:  conf.OAuthGoogleCallbackURL,
		Scopes:       []string{oidc.ScopeOpenID, "email"},
		Endpoint:     googleProvider.Endpoint(),
	})
	if err != nil {
		return nil, fmt.Errorf("new google client: %v", err)
	}

	sessions, err := newSessionStore(conf.RedisURL)
	if err != nil {
		return nil, err
	}

	a := &Authenticator{
		GoogleClient: googleClient,
		sessions:     sessions,
	}
	return a, nil
}

func GenerateRandomState() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)

	return state, nil
}

func (a *Authenticator) Session(r *http.Request) (*Session, error) {
	return a.sessions.GetSession(r)
}
