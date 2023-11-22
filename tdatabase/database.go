package tdatabase

import (
	"context"
)

type DBOperations interface {
	Add(ctx context.Context, in interface{}) error
	GetOne(ctx context.Context, in interface{}, params ParamRequest) (interface{}, error)
	GetMany(ctx context.Context, in interface{}, params ParamRequest) ([]interface{}, error)
	Update(ctx context.Context, update interface{}, in interface{}, id int) error
	Delete(ctx context.Context, in interface{}, id int) error
	Migrate() error
}

type ParamRequest struct {
	By    any `json:"by"`
	Value any `json:"value"`
	From  any `json:"from"`
	To    any `json:"to"`
}
