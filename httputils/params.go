package httputils

import (
	"context"
)

// GetRequestParams : get request params
func GetRequestParams(ctx context.Context) map[string]interface{} {
	m, ok := ctx.Value("params").(map[string]interface{})

	if !ok {
		return map[string]interface{}{}
	}
	return m
}
