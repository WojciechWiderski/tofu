package thttp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/WojciechWiderski/tofu/tcontext"
	"github.com/WojciechWiderski/tofu/terror"
	"github.com/WojciechWiderski/tofu/tlogger"

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
	if db == nil {
		tlogger.Error("DB cannot be nil")
		panic("DB cannot be nil")
	}
	return func(api *HttpAPI) {
		api.Database = db
	}
}

func (a *HttpAPI) GetOne(ctx context.Context, params tdatabase.ParamRequest) (interface{}, error) {

	var err error
	modelFromCtx := tcontext.ModelFromCtx(ctx)

	modelFromCtx.In, err = a.runFn(ctx, tmodel.FnBeforeDBO, modelFromCtx)
	if err != nil {
		return nil, terror.Wrap("a.runFn - FnBeforeDBO", err)
	}

	resp, err := modelFromCtx.Store.GetOne(ctx, modelFromCtx.In, params)
	if err != nil {
		return nil, terror.Wrap(fmt.Sprintf("a.Database.GetOne model - %v by - %v by value - %v.", modelFromCtx, by, params.By), err)
	}

	modelFromCtx.In, err = a.runFn(ctx, tmodel.FnAfterDBO, modelFromCtx)
	if err != nil {
		return nil, terror.Wrap("a.runFn - FnAfterDBO", err)
	}

	return resp, nil
}

func (a *HttpAPI) GetMany(ctx context.Context, params tdatabase.ParamRequest) (interface{}, error) {

	var err error
	modelFromCtx := tcontext.ModelFromCtx(ctx)

	modelFromCtx.In, err = a.runFn(ctx, tmodel.FnBeforeDBO, modelFromCtx)
	if err != nil {
		return nil, terror.Wrap("a.runFn - FnBeforeDBO", err)
	}

	resp, err := modelFromCtx.Store.GetMany(ctx, modelFromCtx.In, params)

	if err != nil {
		return nil, terror.Wrap(fmt.Sprintf("a.Database.GetMany model - %v by - %v by value - %v.", model, by, params.By), err)
	}

	modelFromCtx.In, err = a.runFn(ctx, tmodel.FnAfterDBO, modelFromCtx)
	if err != nil {
		return nil, terror.Wrap("a.runFn - FnAfterDBO", err)
	}

	return resp, nil
}

func (a *HttpAPI) GetOwn(ctx context.Context, w http.ResponseWriter, r *http.Request, params tdatabase.ParamRequest) (interface{}, error) {
	var err error
	modelFromCtx := tcontext.ModelFromCtx(ctx)

	pattern := tcontext.PatternFromCtx(ctx)

	resp, err := modelFromCtx.Routes[http.MethodGet][pattern].Fn(ctx, w, r, a.Database)
	if err != nil {
		return nil, terror.Wrap("%s", err)
	}

	return resp, nil
}

func (a *HttpAPI) AddOne(ctx context.Context, body io.Reader) error {
	var err error
	modelFromCtx := tcontext.ModelFromCtx(ctx)

	if err := json.NewDecoder(body).Decode(&modelFromCtx.In); err != nil {
		return terror.NewInternalf("json.NewDecoder(r.Body)", err)
	}

	modelFromCtx.In, err = a.runFn(ctx, tmodel.FnBeforeDBO, modelFromCtx)
	if err != nil {
		return terror.Wrap("a.runFn - FnBeforeDBO", err)
	}

	if err := modelFromCtx.Store.Add(ctx, modelFromCtx.In); err != nil {
		return terror.Wrap(fmt.Sprintf("a.Database.Add model - %v", model), err)
	}

	modelFromCtx.In, err = a.runFn(ctx, tmodel.FnAfterDBO, modelFromCtx)
	if err != nil {
		return terror.Wrap("a.runFn - FnAfterDBO", err)
	}

	return nil
}

// TODO nie działa.
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
		return terror.Wrap(fmt.Sprintf("a.Database.Add model - %v", model), err)
	}

	model.In, err = a.runFn(ctx, tmodel.FnAfterDBO, model)
	if err != nil {
		return terror.Wrap("a.runFn - FnAfterDBO", err)
	}

	terror.HttpApiHandleSuccess(w, r, http.StatusOK, nil)
	return nil
}

func (a *HttpAPI) Update(ctx context.Context, r *http.Request) (interface{}, error) {

	modelFromCtx := tcontext.ModelFromCtx(ctx)

	update, err := a.getModelFromURL(r)
	if err != nil {
		return nil, terror.Wrap("a.getModelFromURL", err)
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return nil, terror.NewInternalf("strconv.Atoi()", err)
	}

	if err := json.NewDecoder(r.Body).Decode(&update.In); err != nil {
		return nil, terror.NewInternalf("json.NewDecoder(r.Body)", err)
	}

	modelFromCtx.In, err = a.runFn(ctx, tmodel.FnBeforeDBO, modelFromCtx)
	if err != nil {
		return nil, terror.Wrap("a.runFn - FnBeforeDBO", err)
	}

	err = a.Database.Update(ctx, update.In, modelFromCtx.In, id)
	if err != nil {
		return nil, terror.Wrap(fmt.Sprintf("a.Database.Update update - %v model - %v id - %v.", update, modelFromCtx, id), err)
	}

	modelFromCtx.In, err = a.runFn(ctx, tmodel.FnAfterDBO, modelFromCtx)
	if err != nil {
		return nil, terror.Wrap("a.runFn - FnAfterDBO", err)
	}

	return nil, nil
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
		if model.Name == chi.URLParam(r, "model") {
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
