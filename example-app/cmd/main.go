package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/WojciechWiderski/tofu"
	"github.com/WojciechWiderski/tofu/example-app/db"
	"github.com/WojciechWiderski/tofu/example-app/model"
	"github.com/WojciechWiderski/tofu/tconfig"
	"github.com/WojciechWiderski/tofu/tcontext"
	"github.com/WojciechWiderski/tofu/tdatabase"
	"github.com/WojciechWiderski/tofu/terror"
	"github.com/WojciechWiderski/tofu/tmodel"
)

func main() {

	app := tofu.New(tofu.WithHTTPServer(tconfig.HTTP{
		Port: ":3000",
	}, tconfig.Cors{}))
	app.DB = db.New(tconfig.MySql{
		Username:     "root",
		Password:     "PASSWORD",
		Address:      "192.168.1.17",
		DatabaseName: "invest",
	}, app.Models)
	user := tmodel.NewModel(&model.User{}, "user")

	fn := func(ctx context.Context, w http.ResponseWriter, r *http.Request, operations tdatabase.DBOperations) (interface{}, error) {
		m := tcontext.ModelInterfaceFromCtx(ctx).(*model.User)
		req := tdatabase.ParamRequest{
			By:    "id",
			Value: "1",
		}
		resp, err := operations.GetOne(ctx, m, req)
		if err != nil {
			return nil, terror.Wrap(fmt.Sprintf("a.Database.GetOne model - %v by - %v by value - %v.", m, req, req.By), err)
		}

		return resp, err
	}

	user.AddRoute(tmodel.Route{
		RouteType: tmodel.NewRouteType("own"),
		Pattern:   "hi",
		Fn:        fn,
		Method:    http.MethodGet,
	})

	book := tmodel.NewModel(&model.Book{}, "book")

	app.Models.Set(user)
	app.Models.Set(book)

	app.Run()
}
