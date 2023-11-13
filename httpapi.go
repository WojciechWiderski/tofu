package tofu

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type HttpAPI struct {
	Database DBOperations
	Models   *Models
}

type DBOperations interface {
	Add(ctx context.Context, in interface{}) error
	GetOne(ctx context.Context, in interface{}, params ParamRequest) (interface{}, error)
	GetMany(ctx context.Context, in interface{}, params ParamRequest) ([]interface{}, error)
	Update(ctx context.Context, update interface{}, in interface{}, id int) error
	Delete(ctx context.Context, in interface{}, id int) error
	migrate() error
}

const (
	model string = "model"
	id    string = "id"
	by    string = "by"
)

func NewHttpApi(models *Models, opts ...func(*HttpAPI)) *HttpAPI {
	api := &HttpAPI{
		Models: models,
	}

	for _, opt := range opts {
		opt(api)
	}

	return api
}

func WithDatabase(db DBOperations) func(*HttpAPI) {
	return func(api *HttpAPI) {
		api.Database = db
	}
}

func (a *HttpAPI) GetHandler(corsConfig CorsConfig) http.Handler {
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

	for _, model := range a.Models.All {
		r.With(RouteTypeMiddleware(RouteGetOne)).Get(fmt.Sprintf("/{%s}/one", model.Name), ApiHandleError(a.GetOne))
		r.With(RouteTypeMiddleware(RouteGetMany)).Get(fmt.Sprintf("/{%s}/many", model.Name), ApiHandleError(a.GetMany))
		r.With(RouteTypeMiddleware(RouteAddOne)).Post(fmt.Sprintf("/{%s}/add", model.Name), ApiHandleError(a.Add))
		r.With(RouteTypeMiddleware(RouteUpdate)).Put(fmt.Sprintf("/{%s}/{id}/update", model.Name), ApiHandleError(a.Update))
		r.With(RouteTypeMiddleware(RouteDeleteOne)).Delete(fmt.Sprintf("/{%s}/{id}/delete", model.Name), ApiHandleError(a.Delete))

		for _, route := range r.Routes() {
			fmt.Println(route.Pattern)
		}
	}
	return r
}

type ParamRequest struct {
	By    any `json:"by"`
	Value any `json:"value"`
	From  any `json:"from"`
	To    any `json:"to"`
}

func (a *HttpAPI) GetOne(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	model, err := a.getModelFromURL(r)
	if err != nil {
		return Wrap("a.getModelFromURL", err)
	}

	query := r.URL.Query()
	req := ParamRequest{
		By:    query.Get("by"),
		Value: query.Get("value"),
		From:  query.Get("from"),
		To:    query.Get("to"),
	}

	model.In, err = a.runFn(ctx, FnBeforeDBO, model)
	if err != nil {
		return Wrap("a.runFn - FnBeforeDBO", err)
	}

	resp, err := a.Database.GetOne(ctx, model.In, req)
	if err != nil {
		return Wrap(fmt.Sprintf("a.Database.GetOne model - %v by - %v by value - %v.", model, by, req.By), err)
	}

	model.In, err = a.runFn(ctx, FnAfterDBO, model)
	if err != nil {
		return Wrap("a.runFn - FnAfterDBO", err)
	}

	HandleSuccess(w, r, http.StatusOK, resp)
	return nil
}

func (a *HttpAPI) GetMany(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	model, err := a.getModelFromURL(r)
	if err != nil {
		return Wrap("a.getModelFromURL", err)
	}

	query := r.URL.Query()
	req := ParamRequest{
		By:    query.Get("by"),
		Value: query.Get("value"),
		From:  query.Get("from"),
		To:    query.Get("to"),
	}

	model.In, err = a.runFn(ctx, FnBeforeDBO, model)
	if err != nil {
		return Wrap("a.runFn - FnBeforeDBO", err)
	}

	resp, err := a.Database.GetMany(ctx, model.In, req)

	if err != nil {
		return Wrap(fmt.Sprintf("a.Database.GetMany model - %v by - %v by value - %v.", model, by, req.By), err)
	}

	model.In, err = a.runFn(ctx, FnAfterDBO, model)
	if err != nil {
		return Wrap("a.runFn - FnAfterDBO", err)
	}

	HandleSuccess(w, r, http.StatusOK, resp)
	return nil
}

func (a *HttpAPI) Add(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	model, err := a.getModelFromURL(r)
	if err != nil {
		return Wrap("a.getModelFromURL", err)
	}

	if err := json.NewDecoder(r.Body).Decode(&model.In); err != nil {
		return NewInternalf("json.NewDecoder(r.Body)", err)
	}

	model.In, err = a.runFn(ctx, FnBeforeDBO, model)
	if err != nil {
		return Wrap("a.runFn - FnBeforeDBO", err)
	}

	if err := a.Database.Add(ctx, model.In); err != nil {
		return Wrap(fmt.Sprintf("a.Database.Add model - %v", model), err)
	}

	model.In, err = a.runFn(ctx, FnAfterDBO, model)
	if err != nil {
		return Wrap("a.runFn - FnAfterDBO", err)
	}

	HandleSuccess(w, r, http.StatusOK, nil)
	return nil
}

func (a *HttpAPI) Update(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	model, err := a.getModelFromURL(r)
	if err != nil {
		return Wrap("a.getModelFromURL", err)
	}

	update, err := a.getModelFromURL(r)
	if err != nil {
		return Wrap("a.getModelFromURL", err)
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return NewInternalf("strconv.Atoi()", err)
	}

	if err := json.NewDecoder(r.Body).Decode(&update.In); err != nil {
		return NewInternalf("json.NewDecoder(r.Body)", err)
	}

	model.In, err = a.runFn(ctx, FnBeforeDBO, model)
	if err != nil {
		return Wrap("a.runFn - FnBeforeDBO", err)
	}

	err = a.Database.Update(ctx, update.In, model.In, id)
	if err != nil {
		return Wrap(fmt.Sprintf("a.Database.Update update - %v model - %v id - %v.", update, model, id), err)
	}

	model.In, err = a.runFn(ctx, FnAfterDBO, model)
	if err != nil {
		return Wrap("a.runFn - FnAfterDBO", err)
	}

	HandleSuccess(w, r, http.StatusOK, nil)
	return nil
}

func (a *HttpAPI) Delete(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	model, err := a.getModelFromURL(r)
	if err != nil {
		return Wrap("a.getModelFromURL", err)
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return NewInternalf("strconv.Atoi()", err)
	}

	model.In, err = a.runFn(ctx, FnBeforeDBO, model)
	if err != nil {
		return Wrap("a.runFn - FnBeforeDBO", err)
	}

	err = a.Database.Delete(ctx, model.In, id)
	if err != nil {
		return Wrap(fmt.Sprintf("a.Database.Delete in - %v id - %v.", model, id), err)
	}

	model.In, err = a.runFn(ctx, FnAfterDBO, model)
	if err != nil {
		return Wrap("a.runFn - FnAfterDBO", err)
	}

	HandleSuccess(w, r, http.StatusNoContent, nil)
	return nil
}

func (a *HttpAPI) getModelFromURL(r *http.Request) (*Model, error) {
	for _, model := range a.Models.All {
		if model.Name == strings.Split(r.URL.String(), "/")[1] {
			newModelIn := reflect.New(reflect.ValueOf(model.In).Elem().Type()).Interface()

			return &Model{
				In:        newModelIn,
				Name:      model.Name,
				Functions: model.Functions,
				Routes:    model.Routes,
			}, nil
		}
	}
	return nil, nil
}

func (a *HttpAPI) runFn(ctx context.Context, fnType FunctionType, model *Model) (interface{}, error) {
	routeType := RouteTypeFromCtx(ctx)

	allFn, ok := model.Functions[RouteAll]
	if ok {
		if allFn.functionType == fnType {
			f, err := allFn.f(ctx, model.In, a.Database)
			if err != nil {
				return nil, Wrap("allFn.f", err)
			}
			model.In = f
		}
	}

	routeTypeFn, ok := model.Functions[routeType]
	if !ok {
		return model.In, nil
	}

	if allFn.functionType == fnType {
		return routeTypeFn.f(ctx, model.In, a.Database)
	}
	return model.In, nil
}
