package tofu

import "context"

const RouteTypeCtxKey = "route-type-ctx-key"

func RouteTypeFromCtx(ctx context.Context) RouteType {
	if value, ok := ctx.Value(RouteTypeCtxKey).(RouteType); ok {
		return value
	}
	return WrongRtType
}

func ContextWithRouteType(ctx context.Context, routeType RouteType) context.Context {
	return context.WithValue(ctx, RouteTypeCtxKey, routeType)
}
