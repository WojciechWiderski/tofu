package tofu

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func ModelMiddleware(models *Models) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			for _, model := range models.All {

				modelName := chi.URLParam(r, "model")
				if model.Name == modelName {
					r = r.WithContext(context.WithValue(ctx, ModelCtxKey, model))
					next.ServeHTTP(w, r)
					return
				}
			}

			HandleError(w, r, NewBadRequest("wrong path - model"))
		})
	}
}

func RouteTypeMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			routeType := NewRouteType(chi.URLParam(r, "route-type"))

			if routeType != WrongRtType {
				r = r.WithContext(context.WithValue(ctx, RouteTypeCtxKey, routeType))
				next.ServeHTTP(w, r)
				return
			}

			HandleError(w, r, NewBadRequest("wrong path - route type"))
		})
	}
}

func PatternMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			r = r.WithContext(context.WithValue(ctx, PatternCtxKey, chi.URLParam(r, "pattern")))
			next.ServeHTTP(w, r)
		})
	}
}
