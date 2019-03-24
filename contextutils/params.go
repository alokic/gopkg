package contextutils

import "context"

type key string

var (
	Params = key("params")
)

// SetParams set request params
func SetParams(ctx context.Context, val map[string]interface{}) context.Context {
	h, ok := ctx.Value(Params).(map[string]interface{})

	if !ok {
		h = val
	} else {
		for k, v := range val {
			h[k] = v
		}
	}

	return context.WithValue(ctx, Params, h)
}

// GetParams get request params
func GetParams(ctx context.Context) map[string]interface{} {
	m, ok := ctx.Value(Params).(map[string]interface{})

	if !ok {
		return map[string]interface{}{}
	}
	return m
}
