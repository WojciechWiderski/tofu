package tofu

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-chi/chi/v5"
)

type Fn func(ctx context.Context, in interface{}) (interface{}, error)

type Model struct {
	Name     string
	In       interface{}
	Function Fn
	Routes   []chi.Route
}

func NewModel(in interface{}, name string) *Model {
	return &Model{
		Name: name,
		In:   in,
	}
}

func (m *Model) SetFunc(f func(ctx context.Context, in interface{}) (interface{}, error)) *Model {
	m.Function = f
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
