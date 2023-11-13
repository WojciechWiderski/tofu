package tofu

import (
	"context"
	"net/http"
)

func RouteTypeMiddleware(routeType RouteType) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			r = r.WithContext(context.WithValue(ctx, RouteTypeCtxKey, routeType))
			next.ServeHTTP(w, r)
		})
	}
}
