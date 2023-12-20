package thttp

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/WojciechWiderski/tofu/tconfig"
	"github.com/WojciechWiderski/tofu/tcontext"
	"github.com/WojciechWiderski/tofu/tdatabase"
	"github.com/WojciechWiderski/tofu/terror"
	"github.com/WojciechWiderski/tofu/tmodel"
)

func (a *HttpAPI) GetHandler(corsConfig tconfig.Cors) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   corsConfig.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: corsConfig.AllowCredentials,
		MaxAge:           300,
	}))

	r.Route("/api", func(r chi.Router) {
		r.With(ModelMiddleware(a.Models), RouteTypeMiddleware()).Route("/{model}/{route-type}", func(r chi.Router) {
			r.With(PatternMiddleware()).Get("/{pattern}", terror.HttpApiHandleError(a.HandlerGet))
			r.With(PatternMiddleware()).Get("/", terror.HttpApiHandleError(a.HandlerGet))
			r.With(PatternMiddleware()).Post("/{pattern}", terror.HttpApiHandleError(a.HandlerPost))
			r.With(PatternMiddleware()).Post("/", terror.HttpApiHandleError(a.HandlerPost))
			r.With(PatternMiddleware()).Put("/{pattern}", terror.HttpApiHandleError(a.HandlerPut))
			r.With(PatternMiddleware()).Put("/", terror.HttpApiHandleError(a.HandlerPut))
			//r.With(PatternMiddleware()).Delete("/{pattern}", terror.HttpApiHandleError(a.HandlerDelete))
			//r.With(PatternMiddleware()).Delete("/", terror.HttpApiHandleError(a.HandlerDelete))
		})
	})

	return r
}

func (a *HttpAPI) HandlerGet(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ctx := r.Context()
	var resp interface{}

	model, err := a.getModelFromURL(r)
	if err != nil {
		return nil, terror.Wrap("a.getModelFromURL", err)
	}
	ctx = tcontext.ContextWithModel(ctx, model)

	query := r.URL.Query()
	params := tdatabase.ParamRequest{
		By:    query.Get("by"),
		Value: query.Get("value"),
		From:  query.Get("from"),
		To:    query.Get("to"),
	}

	fn := tcontext.RouteTypeFromCtx(ctx)
	switch fn {
	case tmodel.RouteGetOne:
		resp, err = a.GetOne(ctx, params)
	case tmodel.RouteGetMany:
		resp, err = a.GetMany(ctx, params)
	case tmodel.RouteOwn:
		resp, err = a.GetOwn(ctx, w, r, params)
	default:
		return nil, nil
	}
	return resp, err
}

func (a *HttpAPI) HandlerPost(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ctx := r.Context()

	model, err := a.getModelFromURL(r)
	if err != nil {
		return nil, terror.Wrap("a.getModelFromURL", err)
	}
	ctx = tcontext.ContextWithModel(ctx, model)

	fn := tcontext.RouteTypeFromCtx(ctx)
	switch fn {
	case tmodel.RouteAddOne:
		err = a.AddOne(ctx, r.Body)
	case tmodel.RouteAddMany:
		err = a.AddMany(w, r)
	case tmodel.RouteOwn:
	default:
		return nil, nil
	}
	return nil, nil
}

func (a *HttpAPI) HandlerPut(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ctx := r.Context()
	var resp interface{}

	model, err := a.getModelFromURL(r)
	if err != nil {
		return nil, terror.Wrap("a.getModelFromURL", err)
	}
	ctx = tcontext.ContextWithModel(ctx, model)

	fn := tcontext.RouteTypeFromCtx(r.Context())
	switch fn {
	case tmodel.RouteUpdate:
		resp, err = a.Update(ctx, r)
	default:
		return nil, nil
	}
	return resp, err
}

func (a *HttpAPI) HandlerDelete(w http.ResponseWriter, r *http.Request) error {
	fn := tcontext.RouteTypeFromCtx(r.Context())
	switch fn {
	case tmodel.RouteDeleteOne:
		return a.DeleteMany(w, r)
	case tmodel.RouteDeleteMany:
		return a.DeleteMany(w, r)

	default:
		return nil
	}
}
