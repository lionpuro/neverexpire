package auth

import (
	"context"
	"fmt"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Service struct {
	GoogleClient *Client
	sessions     *SessionStore
}

func NewService() (*Service, error) {
	googleProvider, err := oidc.NewProvider(context.Background(), "https://accounts.google.com")
	if err != nil {
		return nil, fmt.Errorf("new google provider: %v", err)
	}
	googleClient, err := newAuthClient(googleProvider, &oauth2.Config{
		ClientID:     os.Getenv("OAUTH_GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH_GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH_GOOGLE_CALLBACK_URL"),
		Scopes:       []string{oidc.ScopeOpenID, "email"},
		Endpoint:     googleProvider.Endpoint(),
	})
	if err != nil {
		return nil, fmt.Errorf("new google client: %v", err)
	}

	sessions, err := newSessionStore()
	if err != nil {
		return nil, err
	}

	s := &Service{
		GoogleClient: googleClient,
		sessions:     sessions,
	}
	return s, nil
}

type Client struct {
	config       *oauth2.Config
	oidcProvider *oidc.Provider
}

func newAuthClient(provider *oidc.Provider, conf *oauth2.Config) (*Client, error) {
	client := &Client{
		config:       conf,
		oidcProvider: provider,
	}
	return client, nil
}

func (a *Client) verifyToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	rawToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("missing field id_token in oauth2 token")
	}

	conf := &oidc.Config{
		ClientID: a.config.ClientID,
	}

	return a.oidcProvider.Verifier(conf).Verify(ctx, rawToken)
}
