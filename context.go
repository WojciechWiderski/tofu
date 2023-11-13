package tofu

import "context"

const (
	RouteTypeCtxKey = "route-type-ctx-key"
	ModelInCtxKey   = "model-ctx-key"
)

func RouteTypeFromCtx(ctx context.Context) RouteType {
	if value, ok := ctx.Value(RouteTypeCtxKey).(RouteType); ok {
		return value
	}
	return WrongRtType
}

func ContextWithRouteType(ctx context.Context, routeType RouteType) context.Context {
	return context.WithValue(ctx, RouteTypeCtxKey, routeType)
}

func ModelInFromCtx(ctx context.Context) interface{} {
	if value, ok := ctx.Value(ModelInCtxKey).(interface{}); ok {
		return value
	}
	return nil
}

func ContextWithModelIn(ctx context.Context, modelIn interface{}) context.Context {
	return context.WithValue(ctx, ModelInCtxKey, modelIn)
}
