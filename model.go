package tofu

import (
	"context"

	"github.com/go-chi/chi/v5"
)

type RouteType uint8

const (
	WrongRtType RouteType = iota
	RouteAll
	RouteGetOne
	RouteGetMany
	RouteAddOne
	RouteAddMany
	RouteUpdate
	RouteDeleteOne
	RouteDeleteMany
)

type FunctionType uint8

const (
	WrongFnType FunctionType = iota
	FnBeforeDBO
	FnAfterDBO
)

type Fn struct {
	functionType FunctionType
	routeType    RouteType
	f            func(ctx context.Context, in interface{}, operations DBOperations) (interface{}, error)
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

func (m *Model) AddFunc(routeType RouteType, functionType FunctionType, f func(ctx context.Context, in interface{}, operations DBOperations) (interface{}, error)) *Model {
	m.Functions[routeType] = Fn{
		functionType: functionType,
		routeType:    routeType,
		f:            f,
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

func (m *Models) Set(model *Model) {
	m.All = append(m.All, model)
}
