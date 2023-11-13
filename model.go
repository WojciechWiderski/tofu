package tofu

import (
	"context"
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
	f            func(ctx context.Context, operations DBOperations) (interface{}, error)
}

type Model struct {
	Name      string
	In        interface{}
	Functions map[RouteType]Fn
	Routes    []Route
}

type Route struct {
	RType RouteType
}

func NewModel(in interface{}, name string) *Model {
	return &Model{
		Name:      name,
		In:        in,
		Functions: make(map[RouteType]Fn),
	}
}

func (m *Model) AddFunc(routeType RouteType, functionType FunctionType, f func(ctx context.Context, operations DBOperations) (interface{}, error)) *Model {
	m.Functions[routeType] = Fn{
		functionType: functionType,
		routeType:    routeType,
		f:            f,
	}
	return m
}
func (m *Model) SetRoute(rtype RouteType) *Model {
	m.Routes = append(m.Routes, Route{RType: rtype})
	return m
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
