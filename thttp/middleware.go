package thttp

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/WojciechWiderski/tofu/tcontext"
	"github.com/WojciechWiderski/tofu/terror"
	"github.com/WojciechWiderski/tofu/tmodel"
)

func ModelMiddleware(models *tmodel.Models) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			for _, model := range models.All {

				modelName := chi.URLParam(r, "tmodel")
				if model.Name == modelName {
					r = r.WithContext(context.WithValue(ctx, tcontext.ModelCtxKey, model))
					next.ServeHTTP(w, r)
					return
				}
			}

			terror.HandleError(w, r, terror.NewBadRequest("wrong path - tmodel"))
		})
	}
}

func RouteTypeMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			routeType := tmodel.NewRouteType(chi.URLParam(r, "route-type"))

			if routeType != tmodel.WrongRtType {
				r = r.WithContext(context.WithValue(ctx, tcontext.RouteTypeCtxKey, routeType))
				next.ServeHTTP(w, r)
				return
			}

			terror.HandleError(w, r, terror.NewBadRequest("wrong path - route type"))
		})
	}
}

func PatternMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			r = r.WithContext(context.WithValue(ctx, tcontext.PatternCtxKey, chi.URLParam(r, "pattern")))
			next.ServeHTTP(w, r)
		})
	}
}
