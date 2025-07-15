package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/lionpuro/neverexpire/hosts"
	"github.com/lionpuro/neverexpire/keys"
	"github.com/lionpuro/neverexpire/logging"
	"github.com/lionpuro/neverexpire/users"
)

type API struct {
	huma     huma.API
	services services
	logger   logging.Logger
}

type services struct {
	users *users.Service
	hosts *hosts.Service
	keys  *keys.Service
}

func New(mux *http.ServeMux, logger logging.Logger, u *users.Service, h *hosts.Service, k *keys.Service) *API {
	conf := huma.DefaultConfig("Host management API", "1.0.0")
	conf.DocsPath = "/docs"
	api := humago.NewWithPrefix(mux, "/api", conf)

	services := services{
		users: u,
		hosts: h,
		keys:  k,
	}

	a := &API{
		huma:     api,
		services: services,
		logger:   logger,
	}
	return a
}

func (a *API) Register() {
	huma.Register(a.huma, huma.Operation{
		OperationID: "get-hosts",
		Method:      http.MethodGet,
		Path:        "/hosts",
		Description: "List tracked hosts",
	}, a.ListHosts)
	huma.Register(a.huma, huma.Operation{
		OperationID: "get-host",
		Method:      http.MethodGet,
		Path:        "/hosts/{name}",
		Description: "Get host by name",
	}, a.GetHost)
	huma.Register(a.huma, huma.Operation{
		OperationID: "create-host",
		Method:      http.MethodPost,
		Path:        "/hosts",
		Description: "Add host",
	}, a.CreateHost)
	huma.Register(a.huma, huma.Operation{
		OperationID: "delete-host",
		Method:      http.MethodDelete,
		Path:        "/hosts/{name}",
		Description: "Delete host",
	}, a.DeleteHost)
}

type Response[T any] struct {
	Body struct {
		Data T `json:"data"`
	}
}

func newResponse[T any](data T) *Response[T] {
	r := &Response[T]{}
	r.Body.Data = data
	return r
}

type CommonInput struct {
	AccessKey string `query:"access_key" required:"true" doc:"API access key required for authentication"`
}

func (a *API) Authenticate(ctx context.Context, rawkey string) (uid string, err error) {
	if len(rawkey) != 128 {
		return "", fmt.Errorf("invalid access key")
	}
	id := rawkey[:8]
	key, err := a.services.keys.ByID(ctx, id)
	if err != nil {
		return "", err
	}
	match := keys.CompareAccessKey(rawkey, key.Hash)
	if !match {
		return "", fmt.Errorf("access keys don't match")
	}
	return key.UserID, nil
}
