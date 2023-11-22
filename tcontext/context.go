package tcontext

import (
	"context"

	"github.com/WojciechWiderski/tofu/tmodel"
)

const (
	RouteTypeCtxKey = "route-type-ctx-key"
	ModelCtxKey     = "model-ctx-key"
	PatternCtxKey   = "pattern-ctx-key"
)

func ContextWithPattern(ctx context.Context, pattern string) context.Context {
	return context.WithValue(ctx, PatternCtxKey, pattern)
}

func ContextWithRouteType(ctx context.Context, routeType tmodel.RouteType) context.Context {
	return context.WithValue(ctx, RouteTypeCtxKey, routeType)
}

func ContextWithModel(ctx context.Context, model *tmodel.Model) context.Context {
	return context.WithValue(ctx, ModelCtxKey, model)
}

func ModelFromCtx(ctx context.Context) *tmodel.Model {
	if value, ok := ctx.Value(ModelCtxKey).(*tmodel.Model); ok {
		return value
	}
	return nil
}

func ModelInterfaceFromCtx(ctx context.Context) interface{} {
	if value, ok := ctx.Value(ModelCtxKey).(*tmodel.Model); ok {
		return value.In
	}
	return nil
}

func RouteTypeFromCtx(ctx context.Context) tmodel.RouteType {
	if value, ok := ctx.Value(RouteTypeCtxKey).(tmodel.RouteType); ok {
		return value
	}
	return tmodel.WrongRtType
}

func PatternFromCtx(ctx context.Context) string {
	if value, ok := ctx.Value(PatternCtxKey).(string); ok {
		return value
	}
	return ""
}
