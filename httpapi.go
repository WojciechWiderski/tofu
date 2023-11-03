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
	Migrate() error
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
		r.With().Post(fmt.Sprintf("/{%s}/one", model.Name), ApiHandleError(a.GetOne))
		r.With().Post(fmt.Sprintf("/{%s}/many", model.Name), ApiHandleError(a.GetMany))
		r.With().Post(fmt.Sprintf("/{%s}/add", model.Name), ApiHandleError(a.Add))
		r.With().Put(fmt.Sprintf("/{%s}/{id}/update", model.Name), ApiHandleError(a.Update))

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
	in, err := a.GetInterfaceFromURL(r)
	if err != nil {
		return Wrap("budget.GetInterfaceFromParam", err)
	}

	query := r.URL.Query()
	req := ParamRequest{
		By:    query.Get("by"),
		Value: query.Get("value"),
		From:  query.Get("from"),
		To:    query.Get("to"),
	}

	if in.Function != nil {
		in.In, err = in.Function(ctx, in.In)
		if err != nil {
			return Wrap("in.Function", err)
		}
	}

	resp, err := a.Database.GetOne(ctx, in.In, req)

	if err != nil {
		return Wrap(fmt.Sprintf("a.DB.Get in - %v by - %v by value - %v.", in, by, req.By), err)
	}

	HandleSuccess(w, r, http.StatusOK, resp)
	return nil
}

func (a *HttpAPI) GetMany(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	in, err := a.GetInterfaceFromURL(r)
	if err != nil {
		return Wrap("budget.GetInterfaceFromParam", err)
	}

	query := r.URL.Query()
	req := ParamRequest{
		By:    query.Get("by"),
		Value: query.Get("value"),
		From:  query.Get("from"),
		To:    query.Get("to"),
	}

	if in.Function != nil {
		in.In, err = in.Function(ctx, in.In)
		if err != nil {
			return Wrap("in.Function", err)
		}
	}

	resp, err := a.Database.GetMany(ctx, in.In, req)

	if err != nil {
		return Wrap(fmt.Sprintf("a.DB.Get in - %v by - %v by value - %v.", in, by, req.By), err)
	}

	HandleSuccess(w, r, http.StatusOK, resp)
	return nil
}

func (a *HttpAPI) Add(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	in, err := a.GetInterfaceFromURL(r)

	if err != nil {
		return Wrap("GetInterfaceFromParam", err)
	}

	if err := json.NewDecoder(r.Body).Decode(&in.In); err != nil {
		return NewInternalf("json.NewDecoder(r.Body)", err)
	}

	if in.Function != nil {
		in.In, err = in.Function(ctx, in.In)
		if err != nil {
			return Wrap("in.Function", err)
		}
	}

	if err := a.Database.Add(ctx, in.In); err != nil {
		return Wrap(fmt.Sprintf("a.DB.Add in - %v", in), err)
	}

	HandleSuccess(w, r, http.StatusOK, nil)
	return nil
}

func (a *HttpAPI) Update(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	in, err := a.GetInterfaceFromURL(r)

	if err != nil {
		return Wrap("budget.GetInterfaceFromParam", err)
	}
	update, err := a.GetInterfaceFromURL(r)

	if err != nil {
		return Wrap("budget.GetInterfaceFromParam", err)
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return NewInternalf("strconv.Atoi()", err)
	}

	if err := json.NewDecoder(r.Body).Decode(&update.In); err != nil {
		return NewInternalf("json.NewDecoder(r.Body)", err)
	}

	if in.Function != nil {
		in.In, err = in.Function(ctx, in.In)
		if err != nil {
			return Wrap("in.Function", err)
		}
	}

	err = a.Database.Update(ctx, update.In, in.In, id)

	if err != nil {
		return Wrap(fmt.Sprintf("a.DB.Update update - %v in - %v id - %v.", update, in, id), err)
	}

	HandleSuccess(w, r, http.StatusOK, nil)
	return nil
}

func (a *HttpAPI) GetInterfaceFromURL(r *http.Request) (*Model, error) {
	for _, model := range a.Models.All {
		if model.Name == strings.Split(r.URL.String(), "/")[1] {
			newModelIn := reflect.New(reflect.ValueOf(model.In).Elem().Type()).Interface()

			return &Model{
				In:       newModelIn,
				Name:     model.Name,
				Function: model.Function,
			}, nil
		}
	}
	return nil, nil
}
