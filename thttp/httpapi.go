package thttp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/WojciechWiderski/tofu/tcontext"
	"github.com/WojciechWiderski/tofu/terror"

	"github.com/WojciechWiderski/tofu/tconfig"
	"github.com/WojciechWiderski/tofu/tdatabase"
	"github.com/WojciechWiderski/tofu/tmodel"
)

type HttpAPI struct {
	Database tdatabase.DBOperations
	Models   *tmodel.Models
}

const (
	model string = "model"
	id    string = "id"
	by    string = "by"
)

func NewHttpApi(models *tmodel.Models, opts ...func(*HttpAPI)) *HttpAPI {
	api := &HttpAPI{
		Models: models,
	}

	for _, opt := range opts {
		opt(api)
	}

	return api
}

func WithDatabase(db tdatabase.DBOperations) func(*HttpAPI) {
	return func(api *HttpAPI) {
		api.Database = db
	}
}

func (a *HttpAPI) GetHandler(corsConfig tconfig.Cors) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   corsConfig.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: corsConfig.AllowCredentials,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Route("/api", func(r chi.Router) {
		r.With(ModelMiddleware(a.Models), RouteTypeMiddleware()).Route("/{tmodel}/{route-type}", func(r chi.Router) {
			r.With(PatternMiddleware()).Get("/{pattern}", terror.HttpApiHandleError(a.HandlerGet))
			r.With(PatternMiddleware()).Get("/", terror.HttpApiHandleError(a.HandlerGet))
			r.With(PatternMiddleware()).Post("/{pattern}", terror.HttpApiHandleError(a.HandlerPost))
			r.With(PatternMiddleware()).Post("/", terror.HttpApiHandleError(a.HandlerPost))
			r.With(PatternMiddleware()).Put("/{pattern}", terror.HttpApiHandleError(a.HandlerPut))
			r.With(PatternMiddleware()).Put("/", terror.HttpApiHandleError(a.HandlerPut))
			r.With(PatternMiddleware()).Delete("/{pattern}", terror.HttpApiHandleError(a.HandlerDelete))
			r.With(PatternMiddleware()).Delete("/", terror.HttpApiHandleError(a.HandlerDelete))
		})
	})

	return r
}

func (a *HttpAPI) HandlerGet(w http.ResponseWriter, r *http.Request) error {
	fn := tcontext.RouteTypeFromCtx(r.Context())
	switch fn {
	case tmodel.RouteGetOne:
		return a.GetOne(w, r)
	case tmodel.RouteGetMany:
		return a.GetMany(w, r)
	case tmodel.RouteOwn:
		return a.GetOwn(w, r)
	default:
		return nil
	}
}

func (a *HttpAPI) HandlerPost(w http.ResponseWriter, r *http.Request) error {
	fn := tcontext.RouteTypeFromCtx(r.Context())
	switch fn {
	case tmodel.RouteAddOne:
		return a.AddOne(w, r)
	case tmodel.RouteAddMany:
		return a.AddMany(w, r)
	default:
		return nil
	}
}

func (a *HttpAPI) HandlerPut(w http.ResponseWriter, r *http.Request) error {
	fn := tcontext.RouteTypeFromCtx(r.Context())
	switch fn {
	case tmodel.RouteUpdate:
		return a.Update(w, r)
	default:
		return nil
	}
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

func (a *HttpAPI) GetOne(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	model, err := a.getModelFromURL(r)
	if err != nil {
		return terror.Wrap("a.getModelFromURL", err)
	}

	query := r.URL.Query()
	req := tdatabase.ParamRequest{
		By:    query.Get("by"),
		Value: query.Get("value"),
		From:  query.Get("from"),
		To:    query.Get("to"),
	}

	model.In, err = a.runFn(ctx, tmodel.FnBeforeDBO, model)
	if err != nil {
		return terror.Wrap("a.runFn - FnBeforeDBO", err)
	}

	resp, err := a.Database.GetOne(ctx, model.In, req)
	if err != nil {
		return terror.Wrap(fmt.Sprintf("a.Database.GetOne tmodel - %v by - %v by value - %v.", model, by, req.By), err)
	}

	model.In, err = a.runFn(ctx, tmodel.FnAfterDBO, model)
	if err != nil {
		return terror.Wrap("a.runFn - FnAfterDBO", err)
	}

	terror.HttpApiHandleSuccess(w, r, http.StatusOK, resp)
	return nil
}

func (a *HttpAPI) GetMany(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	model, err := a.getModelFromURL(r)
	if err != nil {
		return terror.Wrap("a.getModelFromURL", err)
	}

	query := r.URL.Query()
	req := tdatabase.ParamRequest{
		By:    query.Get("by"),
		Value: query.Get("value"),
		From:  query.Get("from"),
		To:    query.Get("to"),
	}

	model.In, err = a.runFn(ctx, tmodel.FnBeforeDBO, model)
	if err != nil {
		return terror.Wrap("a.runFn - FnBeforeDBO", err)
	}

	resp, err := a.Database.GetMany(ctx, model.In, req)

	if err != nil {
		return terror.Wrap(fmt.Sprintf("a.Database.GetMany tmodel - %v by - %v by value - %v.", model, by, req.By), err)
	}

	model.In, err = a.runFn(ctx, tmodel.FnAfterDBO, model)
	if err != nil {
		return terror.Wrap("a.runFn - FnAfterDBO", err)
	}

	terror.HttpApiHandleSuccess(w, r, http.StatusOK, resp)
	return nil
}

func (a *HttpAPI) GetOwn(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	model, err := a.getModelFromURL(r)
	if err != nil {
		return terror.Wrap("a.getModelFromURL", err)
	}
	pattern := tcontext.PatternFromCtx(ctx)

	resp, err := model.Routes[http.MethodGet][pattern].Fn(ctx, w, r, a.Database)
	if err != nil {
		return terror.Wrap("%s", err)
	}
	terror.HttpApiHandleSuccess(w, r, http.StatusOK, resp)
	return nil
}

func (a *HttpAPI) AddOne(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	model, err := a.getModelFromURL(r)
	if err != nil {
		return terror.Wrap("a.getModelFromURL", err)
	}

	if err := json.NewDecoder(r.Body).Decode(&model.In); err != nil {
		return terror.NewInternalf("json.NewDecoder(r.Body)", err)
	}

	ctx = tcontext.ContextWithModel(ctx, model)

	model.In, err = a.runFn(ctx, tmodel.FnBeforeDBO, model)
	if err != nil {
		return terror.Wrap("a.runFn - FnBeforeDBO", err)
	}

	if err := a.Database.Add(ctx, model.In); err != nil {
		return terror.Wrap(fmt.Sprintf("a.Database.Add tmodel - %v", model), err)
	}

	model.In, err = a.runFn(ctx, tmodel.FnAfterDBO, model)
	if err != nil {
		return terror.Wrap("a.runFn - FnAfterDBO", err)
	}

	terror.HttpApiHandleSuccess(w, r, http.StatusOK, nil)
	return nil
}

func (a *HttpAPI) AddMany(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	model, err := a.getModelFromURL(r)
	if err != nil {
		return terror.Wrap("a.getModelFromURL", err)
	}

	if err := json.NewDecoder(r.Body).Decode(&model.In); err != nil {
		return terror.NewInternalf("json.NewDecoder(r.Body)", err)
	}

	ctx = tcontext.ContextWithModel(ctx, model)

	model.In, err = a.runFn(ctx, tmodel.FnBeforeDBO, model)
	if err != nil {
		return terror.Wrap("a.runFn - FnBeforeDBO", err)
	}

	if err := a.Database.Add(ctx, model.In); err != nil {
		return terror.Wrap(fmt.Sprintf("a.Database.Add tmodel - %v", model), err)
	}

	model.In, err = a.runFn(ctx, tmodel.FnAfterDBO, model)
	if err != nil {
		return terror.Wrap("a.runFn - FnAfterDBO", err)
	}

	terror.HttpApiHandleSuccess(w, r, http.StatusOK, nil)
	return nil
}

func (a *HttpAPI) Update(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	model, err := a.getModelFromURL(r)
	if err != nil {
		return terror.Wrap("a.getModelFromURL", err)
	}

	update, err := a.getModelFromURL(r)
	if err != nil {
		return terror.Wrap("a.getModelFromURL", err)
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return terror.NewInternalf("strconv.Atoi()", err)
	}

	if err := json.NewDecoder(r.Body).Decode(&update.In); err != nil {
		return terror.NewInternalf("json.NewDecoder(r.Body)", err)
	}

	model.In, err = a.runFn(ctx, tmodel.FnBeforeDBO, model)
	if err != nil {
		return terror.Wrap("a.runFn - FnBeforeDBO", err)
	}

	err = a.Database.Update(ctx, update.In, model.In, id)
	if err != nil {
		return terror.Wrap(fmt.Sprintf("a.Database.Update update - %v tmodel - %v id - %v.", update, model, id), err)
	}

	model.In, err = a.runFn(ctx, tmodel.FnAfterDBO, model)
	if err != nil {
		return terror.Wrap("a.runFn - FnAfterDBO", err)
	}

	terror.HttpApiHandleSuccess(w, r, http.StatusOK, nil)
	return nil
}

func (a *HttpAPI) DeleteOne(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	model, err := a.getModelFromURL(r)
	if err != nil {
		return terror.Wrap("a.getModelFromURL", err)
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return terror.NewInternalf("strconv.Atoi()", err)
	}

	model.In, err = a.runFn(ctx, tmodel.FnBeforeDBO, model)
	if err != nil {
		return terror.Wrap("a.runFn - FnBeforeDBO", err)
	}

	err = a.Database.Delete(ctx, model.In, id)
	if err != nil {
		return terror.Wrap(fmt.Sprintf("a.Database.Delete in - %v id - %v.", model, id), err)
	}

	model.In, err = a.runFn(ctx, tmodel.FnAfterDBO, model)
	if err != nil {
		return terror.Wrap("a.runFn - FnAfterDBO", err)
	}

	terror.HttpApiHandleSuccess(w, r, http.StatusNoContent, nil)
	return nil
}
func (a *HttpAPI) DeleteMany(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	model, err := a.getModelFromURL(r)
	if err != nil {
		return terror.Wrap("a.getModelFromURL", err)
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return terror.NewInternalf("strconv.Atoi()", err)
	}

	model.In, err = a.runFn(ctx, tmodel.FnBeforeDBO, model)
	if err != nil {
		return terror.Wrap("a.runFn - FnBeforeDBO", err)
	}

	err = a.Database.Delete(ctx, model.In, id)
	if err != nil {
		return terror.Wrap(fmt.Sprintf("a.Database.Delete in - %v id - %v.", model, id), err)
	}

	model.In, err = a.runFn(ctx, tmodel.FnAfterDBO, model)
	if err != nil {
		return terror.Wrap("a.runFn - FnAfterDBO", err)
	}

	terror.HttpApiHandleSuccess(w, r, http.StatusNoContent, nil)
	return nil
}

func (a *HttpAPI) getModelFromURL(r *http.Request) (*tmodel.Model, error) {
	for _, model := range a.Models.All {
		if model.Name == chi.URLParam(r, "tmodel") {
			newModelIn := reflect.New(reflect.ValueOf(model.In).Elem().Type()).Interface()

			return &tmodel.Model{
				In:        newModelIn,
				Name:      model.Name,
				Functions: model.Functions,
				Routes:    model.Routes,
			}, nil
		}
	}
	return nil, nil
}

func (a *HttpAPI) runFn(ctx context.Context, fnType tmodel.FunctionType, model *tmodel.Model) (interface{}, error) {
	routeType := tcontext.RouteTypeFromCtx(ctx)

	allFn, ok := model.Functions[tmodel.RouteAll]
	if ok {
		if allFn.FunctionType == fnType {
			f, err := allFn.F(ctx, a.Database)
			if err != nil {
				return nil, terror.Wrap("allFn.f", err)
			}
			model.In = f
		}
	}

	routeTypeFn, ok := model.Functions[routeType]
	if !ok {
		return model.In, nil
	}

	if routeTypeFn.FunctionType == fnType {
		return routeTypeFn.F(ctx, a.Database)
	}
	return model.In, nil
}
