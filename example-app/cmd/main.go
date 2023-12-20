package main

import (
	"github.com/WojciechWiderski/tofu"
	"github.com/WojciechWiderski/tofu/example-app/model"
	"github.com/WojciechWiderski/tofu/tconfig"
	"github.com/WojciechWiderski/tofu/tmodel"
)

func main() {
	app := tofu.New(
		tofu.WithMySQLDB(tconfig.MySql{
			Username:     "user",
			Password:     "password",
			Address:      "localhost:3306",
			DatabaseName: "db",
		}),
	)

	app.Models.Set(tmodel.NewModel(&model.Day{}, "day"))
	app.Models.Set(tmodel.NewModel(&model.Date{}, "date"))
	app.Models.Set(tmodel.NewModel(&model.User{}, "user"))
	app.Models.Set(tmodel.NewModel(&model.Level{}, "level"))
	app.Models.Set(tmodel.NewModel(&model.Task{}, "task"))

	app.Run()
}
