package tmodel

import (
	"context"

	"github.com/WojciechWiderski/tofu/tdatabase"
)

type FunctionType uint8

const (
	WrongFnType FunctionType = iota
	FnBeforeDBO
	FnAfterDBO
)

var FunctionTypeMap = map[string]FunctionType{
	"wrong-fn-type": WrongFnType,
	"fn-before-dbo": FnBeforeDBO,
	"fn-after-dbo":  FnAfterDBO,
}

func (rt FunctionType) String() string {
	switch rt {
	case WrongFnType:
		return "wrong-fn-type"
	case FnBeforeDBO:
		return "fn-before-dbo"
	case FnAfterDBO:
		return "fn-after-dbo"
	default:
		return ""
	}
}

func NewFunctionType(in string) FunctionType {
	if len(in) == 0 {
		return WrongFnType
	}

	if r, ok := FunctionTypeMap[in]; ok {
		return r
	} else {
		return WrongFnType
	}
}

type Fn struct {
	FunctionType FunctionType
	RouteType    RouteType
	F            func(ctx context.Context, operations tdatabase.DBOperations) (interface{}, error)
}

func NewFn(fnType FunctionType, rType RouteType, f func(ctx context.Context, operations tdatabase.DBOperations) (interface{}, error)) Fn {
	return Fn{
		FunctionType: fnType,
		RouteType:    rType,
		F:            f,
	}
}
