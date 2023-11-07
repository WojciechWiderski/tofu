package tofu

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type RouteType uint8

const (
	RouteTypeCtxKey = "route-type-ctx-key"
)

const (
	WrongType RouteType = iota
	GetOne
	GetMany
	AddOne
	AddMany
	Update
	DeleteOne
	DeleteMany
)

type Fn struct {
	routeType RouteType
	f         func(ctx context.Context, in interface{}) (interface{}, error)
}

func RouteTypeFromCtx(ctx context.Context) RouteType {
	if value, ok := ctx.Value(RouteTypeCtxKey).(RouteType); ok {
		return value
	}
	return WrongType
}
func ContextWithRouteType(ctx context.Context, routeType RouteType) context.Context {
	return context.WithValue(ctx, RouteTypeCtxKey, routeType)
}

func RouteTypeMiddleware(routeType RouteType) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			r = r.WithContext(context.WithValue(ctx, RouteTypeCtxKey, routeType))
			next.ServeHTTP(w, r)
		})
	}
}

type Model struct {
	Name      string
	In        interface{}
	Functions map[RouteType]Fn
	Routes    []chi.Route
}

func NewModel(in interface{}, name string) *Model {
	return &Model{
		Name:      name,
		In:        in,
		Functions: make(map[RouteType]Fn),
	}
}

func (m *Model) SetFunc(routeType RouteType, f func(ctx context.Context, in interface{}) (interface{}, error)) *Model {
	m.Functions[routeType] = Fn{
		routeType: routeType,
		f:         f,
	}
	return m
}
func (m *Model) SetRoute(pattern string) *Model {
	return nil
}

type Models struct {
	All []*Model
}

func NewModels(models ...*Model) *Models {
	return &Models{All: models}
}

func (m *Models) GetInterfaces() []interface{} {
	var ins []interface{}
	for _, model := range m.All {
		ins = append(ins, model.In)
	}
	return ins
}

func (m *Models) Set(model *Model) {
	m.All = append(m.All, model)
}

func (m *Models) GetModels() *Models {
	return m
}

func (m *Model) Decode(body io.Reader) {
	if err := json.NewDecoder(body).Decode(&m.In); err != nil {
		fmt.Println(err)
	}
}
