package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/lionpuro/trackcert/model"
	"golang.org/x/oauth2"
)

const (
	googleUserEndpoint = "https://www.googleapis.com/oauth2/v2/userinfo"
	userContextKey     = "user"
)

type AuthService struct {
	GoogleClient *AuthClient
}

type AuthClient struct {
	config       *oauth2.Config
	oidcProvider *oidc.Provider
}

func newAuthService() (*AuthService, error) {
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

	s := &AuthService{
		GoogleClient: googleClient,
	}
	return s, nil
}

func newAuthClient(provider *oidc.Provider, conf *oauth2.Config) (*AuthClient, error) {
	client := &AuthClient{
		config:       conf,
		oidcProvider: provider,
	}
	return client, nil
}

func (s *Server) handleAuth(a *AuthClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state, err := generateRandomState()
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		sess, err := s.Sessions.GetSession(r)
		if err != nil {
			log.Printf("get session: %v", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		sess.Values["state"] = state
		if err := sess.Save(r, w); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		url := a.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func (s *Server) handleAuthCallback(a *AuthClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, err := s.Sessions.GetSession(r)
		if err != nil {
			log.Printf("get session: %v", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		if r.FormValue("state") != sess.Values["state"] {
			http.Error(w, "Invalid state parameter.", http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")
		tkn, err := a.config.Exchange(r.Context(), code)
		if err != nil {
			log.Printf("auth callback: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		idToken, err := a.verifyToken(r.Context(), tkn)
		if err != nil {
			log.Printf("verify token: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		var user struct {
			ID    string `json:"sub"`
			Email string `json:"email"`
		}
		if err := idToken.Claims(&user); err != nil {
			log.Printf("unmarshal claims: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if err := s.DB.CreateUser(user.ID, user.Email); err != nil {
			log.Printf("%v", err)
			http.Error(w, "Error registering user", http.StatusInternalServerError)
			return
		}
		sess.Values["user"] = model.SessionUser{ID: user.ID, Email: user.Email}
		if err := sess.Save(r, w); err != nil {
			log.Printf("save session: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	sess, err := s.Sessions.GetSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sess.Options.MaxAge = -1
	if err := sess.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (a *AuthClient) verifyToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	rawToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("missing field id_token in oauth2 token")
	}

	conf := &oidc.Config{
		ClientID: a.config.ClientID,
	}

	return a.oidcProvider.Verifier(conf).Verify(ctx, rawToken)
}

func generateRandomState() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)

	return state, nil
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := getUserCtx(r.Context())
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}
