package tmodel

import (
	"context"
	"net/http"

	"github.com/WojciechWiderski/tofu/tdatabase"
)

type Route struct {
	RouteType RouteType
	Pattern   string
	Fn        func(ctx context.Context, w http.ResponseWriter, r *http.Request, operations tdatabase.DBOperations) (interface{}, error)
	Method    string
}

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
		return "own"
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

func NewRoute(routeType RouteType, pattern string, method string, fn func(ctx context.Context, w http.ResponseWriter, r *http.Request, operations tdatabase.DBOperations) (interface{}, error)) {
	// sprawdzać czy pattern jest ok i czy toutetype jest ok, przenieiść z set route, albo do set route
}
