package auth

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

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

func (a *Client) VerifyToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	rawToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("missing field id_token in oauth2 token")
	}

	conf := &oidc.Config{
		ClientID: a.config.ClientID,
	}

	return a.oidcProvider.Verifier(conf).Verify(ctx, rawToken)
}

func (a *Client) AuthCodeURL(state string) string {
	return a.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

func (a *Client) ExchangeToken(ctx context.Context, code string) (*oauth2.Token, error) {
	tkn, err := a.config.Exchange(ctx, code)
	return tkn, err
}
