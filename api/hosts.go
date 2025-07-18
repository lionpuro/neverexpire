package api

import (
	"context"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lionpuro/neverexpire/db"
	"github.com/lionpuro/neverexpire/hosts"
	"github.com/lionpuro/neverexpire/users"
)

type Host struct {
	HostName  string     `json:"hostname"`
	Issuer    *string    `json:"issuer"`
	ExpiresAt *time.Time `json:"expires_at"`
	CheckedAt time.Time  `json:"checked_at"`
	Error     *string    `json:"error"`
}

func toAPISchema(h hosts.Host) Host {
	var errMsg *string
	if err := h.Certificate.Error; err != nil {
		msg := err.Error()
		errMsg = &msg
	}
	result := Host{
		HostName:  h.HostName,
		Issuer:    &h.Certificate.IssuedBy,
		ExpiresAt: h.Certificate.ExpiresAt,
		CheckedAt: h.Certificate.CheckedAt,
		Error:     errMsg,
	}
	if iss := h.Certificate.IssuedBy; iss == "n/a" || iss == "" {
		result.Issuer = nil
	}
	return result
}

type HostsInput struct{}

func (a *API) ListHosts(ctx context.Context, input *HostsInput) (*Response[[]Host], error) {
	uid, ok := currentUID(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	hsts, err := a.services.hosts.AllByUser(ctx, uid)
	if err != nil {
		a.logger.Error("failed to get hosts", "error", err.Error())
		return nil, huma.Error500InternalServerError("failed to retrieve hosts")
	}
	var result []Host
	for _, h := range hsts {
		host := toAPISchema(h)
		result = append(result, host)
	}
	return newResponse(result), nil
}

type HostInput struct {
	Name string `path:"name"`
}

func (a *API) GetHost(ctx context.Context, input *HostInput) (*Response[Host], error) {
	uid, ok := currentUID(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	host, err := a.services.hosts.ByName(ctx, input.Name, uid)
	if err != nil {
		if db.IsErrNoRows(err) {
			return nil, huma.Error404NotFound("host not found")
		}
		a.logger.Error("failed to get host", "error", err.Error())
		return nil, huma.Error500InternalServerError("failed to retrieve host information")
	}
	result := toAPISchema(host)
	return newResponse(result), nil
}

type CreateHostInput struct {
	Body struct {
		Name string `json:"name" required:"true"`
	}
}

func (a *API) CreateHost(ctx context.Context, input *CreateHostInput) (*Response[Host], error) {
	uid, ok := currentUID(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	name, err := hosts.ParseHostname(input.Body.Name)
	if err != nil {
		return nil, huma.Error400BadRequest("bad request")
	}
	if err := a.services.hosts.Create(users.User{ID: uid}, []string{name}); err != nil {
		if strings.Contains(err.Error(), "already tracking") {
			host, err := a.services.hosts.ByName(ctx, name, uid)
			if err != nil {
				a.logger.Error("failed to retrieve new host by name", "error", err.Error())
				return nil, huma.Error500InternalServerError("failed to get created host")
			}
			return newResponse(toAPISchema(host)), nil
		}
		return nil, huma.Error500InternalServerError("failed to create host")
	}
	host, err := a.services.hosts.ByName(ctx, name, uid)
	if err != nil {
		a.logger.Error("failed to retrieve new host by name", "error", err.Error())
		return nil, huma.Error500InternalServerError("failed to retrieve information for created host")
	}
	return newResponse(toAPISchema(host)), nil
}

func (a *API) DeleteHost(ctx context.Context, input *HostInput) (*struct{}, error) {
	uid, ok := currentUID(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	host, err := a.services.hosts.ByName(ctx, input.Name, uid)
	if err != nil {
		if db.IsErrNoRows(err) {
			return nil, huma.Error404NotFound("host not found")
		}
		a.logger.Error("failed to get host", "error", err.Error())
		return nil, huma.Error500InternalServerError("failed to delete host")
	}
	if err := a.services.hosts.Delete(uid, host.ID); err != nil {
		a.logger.Error("failed to get host", "error", err.Error())
		return nil, huma.Error500InternalServerError("failed to delete host")
	}
	return nil, nil
}
