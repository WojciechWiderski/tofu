package tofu

import "context"

const (
	RouteTypeCtxKey = "route-type-ctx-key"
	ModelCtxKey     = "model-ctx-key"
	PatternCtxKey   = "pattern-ctx-key"
)

func PatternFromCtx(ctx context.Context) string {
	if value, ok := ctx.Value(PatternCtxKey).(string); ok {
		return value
	}
	return ""
}

func ContextWithPattern(ctx context.Context, pattern string) context.Context {
	return context.WithValue(ctx, PatternCtxKey, pattern)
}

func RouteTypeFromCtx(ctx context.Context) RouteType {
	if value, ok := ctx.Value(RouteTypeCtxKey).(RouteType); ok {
		return value
	}
	return WrongRtType
}

func ContextWithRouteType(ctx context.Context, routeType RouteType) context.Context {
	return context.WithValue(ctx, RouteTypeCtxKey, routeType)
}

func ModelFromCtx(ctx context.Context) *Model {
	if value, ok := ctx.Value(ModelCtxKey).(*Model); ok {
		return value
	}
	return nil
}

func ModelInterfaceFromCtx(ctx context.Context) interface{} {
	if value, ok := ctx.Value(ModelCtxKey).(*Model); ok {
		return value.In
	}
	return nil
}

func ContextWithModel(ctx context.Context, model *Model) context.Context {
	return context.WithValue(ctx, ModelCtxKey, model)
}
