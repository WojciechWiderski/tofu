package tofu

import (
	"context"
	"net/http"
)

type RouteType uint8

const (
	WrongRtType RouteType = iota
	RouteOwn
	RouteAll
	RouteGetOne
	RouteGetMany
	RouteAddOne
	RouteAddMany
	RouteUpdate
	RouteDeleteOne
	RouteDeleteMany
)

var RouteTypeMap = map[string]RouteType{
	"wrong":       WrongRtType,
	"own":         RouteOwn,
	"all":         RouteAll,
	"get-one":     RouteGetOne,
	"get-many":    RouteGetMany,
	"add-one":     RouteAddOne,
	"add-many":    RouteAddMany,
	"update":      RouteUpdate,
	"delete-one":  RouteDeleteOne,
	"delete-many": RouteDeleteMany,
}

func NewRouteType(in string) RouteType {
	if len(in) == 0 {
		return WrongRtType
	}

	if r, ok := RouteTypeMap[in]; ok {
		return r
	} else {
		return RouteOwn
	}
}

func (rt RouteType) String() string {
	switch rt {
	case RouteOwn:
		return "o"
	case RouteAll:
		return "all"
	case RouteGetOne:
		return "get-one"
	case RouteGetMany:
		return "get-many"
	case RouteAddOne:
		return "add-one"
	case RouteAddMany:
		return "add-many"
	case RouteUpdate:
		return "update"
	case RouteDeleteOne:
		return "delete-one"
	case RouteDeleteMany:
		return "delete-many"
	default:
		return ""
	}
}

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
	Routes    map[string]map[string]Route
}

type Route struct {
	RouteType RouteType
	Pattern   string
	Fn        func(ctx context.Context, w http.ResponseWriter, r *http.Request, operations DBOperations) (interface{}, error)
	Method    string
}

func NewRoute(routeType RouteType, pattern string, method string, fn func(ctx context.Context, w http.ResponseWriter, r *http.Request, operations DBOperations) (interface{}, error)) {
	// sprawdzać czy pattern jest ok i czy toutetype jest ok, przenieiść z set route, albo do set route
}

func NewFn(fnType FunctionType, rType RouteType, f func(ctx context.Context, operations DBOperations) (interface{}, error)) Fn {
	return Fn{
		functionType: fnType,
		routeType:    rType,
		f:            f,
	}
}

func NewModel(in interface{}, name string) *Model {
	return &Model{
		Name:      name,
		In:        in,
		Functions: make(map[RouteType]Fn),
		Routes:    make(map[string]map[string]Route),
	}
}

func (m *Model) AddFunc(routeType RouteType, functionType FunctionType, f func(ctx context.Context, operations DBOperations) (interface{}, error)) *Model {
	m.Functions[routeType] = NewFn(functionType, routeType, f)
	return m
}

func (m *Model) SetRoute(route Route) *Model {
	_, ok := m.Routes[route.Method][route.Pattern]
	if !ok {
		m.Routes[route.Method] = make(map[string]Route)
	}
	m.Routes[route.Method][route.Pattern] = route
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
