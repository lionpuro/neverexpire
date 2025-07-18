package api

import (
	"net/http"
	"strings"

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
	conf := huma.DefaultConfig("neverexpire.xyz", "1.0.0")
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
	mw := huma.Middlewares{newAuthMiddleware(a)}
	huma.Register(a.huma, huma.Operation{
		OperationID: "get-hosts",
		Method:      http.MethodGet,
		Path:        "/hosts",
		Description: "List tracked hosts",
		Middlewares: mw,
	}, a.ListHosts)
	huma.Register(a.huma, huma.Operation{
		OperationID: "get-host",
		Method:      http.MethodGet,
		Path:        "/hosts/{name}",
		Description: "Get host by name",
		Middlewares: mw,
	}, a.GetHost)
	huma.Register(a.huma, huma.Operation{
		OperationID: "create-host",
		Method:      http.MethodPost,
		Path:        "/hosts",
		Description: "Add host",
		Middlewares: mw,
	}, a.CreateHost)
	huma.Register(a.huma, huma.Operation{
		OperationID: "delete-host",
		Method:      http.MethodDelete,
		Path:        "/hosts/{name}",
		Description: "Delete host",
		Middlewares: mw,
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

func (a *API) writeErr(ctx huma.Context, status int, msg string, errs ...error) {
	if err := huma.WriteErr(a.huma, ctx, status, msg, errs...); err != nil {
		a.logger.Error("failed to write error", "error", err.Error())
	}
}

func newAuthMiddleware(a *API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		rawkey := strings.TrimPrefix(ctx.Header("Authorization"), "Bearer ")
		if len(rawkey) != 128 {
			a.writeErr(ctx, http.StatusUnauthorized, "unauthorized")
			return
		}
		id := rawkey[:8]
		key, err := a.services.keys.ByID(ctx.Context(), id)
		if err != nil {
			a.writeErr(ctx, http.StatusUnauthorized, "unauthorized")
			return
		}
		match := keys.CompareAccessKey(rawkey, key.Hash)
		if !match {
			a.writeErr(ctx, http.StatusUnauthorized, "unauthorized")
			return
		}
		next(huma.WithValue(ctx, ctxKeyUID, key.UserID))
	}
}
