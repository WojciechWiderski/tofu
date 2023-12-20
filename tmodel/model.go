package tmodel

import (
	"context"
	"fmt"

	"github.com/WojciechWiderski/tofu/tdatabase"
	"github.com/WojciechWiderski/tofu/tlogger"
)

type Model struct {
	Name      string
	In        interface{}
	Functions map[RouteType]Fn
	Routes    map[string]map[string]Route
	DB        tdatabase.DBOperations
}

func NewModel(in interface{}, name string) *Model {
	tlogger.Info(fmt.Sprintf("Model %s created.", name))
	return &Model{
		Name:      name,
		In:        in,
		Functions: make(map[RouteType]Fn),
		Routes:    make(map[string]map[string]Route),
	}
}

func (m *Model) AddFunc(routeType RouteType, functionType FunctionType, f func(ctx context.Context, operations tdatabase.DBOperations) (interface{}, error)) *Model {
	m.Functions[routeType] = NewFn(functionType, routeType, f)
	tlogger.Info(fmt.Sprintf("AddFunc for model - %s, route type: %s, function type: %s", m.Name, routeType.String(), functionType.String()))
	return m
}

func (m *Model) AddRoute(route Route) *Model {
	_, ok := m.Routes[route.Method][route.Pattern]
	if !ok {
		m.Routes[route.Method] = make(map[string]Route)
	}
	m.Routes[route.Method][route.Pattern] = route
	tlogger.Info(fmt.Sprintf("AddRoute for model - %s, route method: %s, route pattern: %s, route type: %s ", m.Name, route.Method, route.Pattern, route.RouteType.String()))
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

func (m *Models) Get(name string) *Model {
	for _, model := range m.All {
		if model.Name == name {
			return model
		}
	}
	return nil
}
