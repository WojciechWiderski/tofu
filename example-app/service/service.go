package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/WojciechWiderski/tofu/example-app/model"
	"github.com/WojciechWiderski/tofu/tdatabase"
	"github.com/WojciechWiderski/tofu/terror"
	"github.com/WojciechWiderski/tofu/tlogger"
	"github.com/WojciechWiderski/tofu/tmodel"
)

type Service struct {
	models *tmodel.Models
}

func New(models *tmodel.Models) *Service {
	return &Service{models: models}
}

func (s *Service) AddUser() error {
	ctx := context.Background()

	userModel, err := s.models.GetRawModel("user")
	if err != nil {
		return terror.Wrap("s.models.GetRawModel", err)
	}

	user, err := userModel.Store.GetOne(ctx, userModel.In, tdatabase.ParamRequest{
		By:    "id",
		Value: "1",
	})
	if err != nil {
		if errors.Is(err, fmt.Errorf("record not found")) {
			fmt.Println("xD")
		}
		return terror.Wrap("userModel.Store.GetOne", err)
	}

	if user != nil {
		tlogger.Info("User already exist")
		return nil
	}

	err = userModel.Store.Add(ctx, model.User{
		CurrentExp:   0,
		CurrentLevel: 0,
	})
	if err != nil {
		return terror.Wrap("userModel.Store.Add", err)
	}

	tlogger.Success("AddUser successful!")
	return nil
}
