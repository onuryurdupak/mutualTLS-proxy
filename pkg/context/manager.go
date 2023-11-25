package context

import (
	"context"
)

type key string

const (
	keyNameSessionID key = "sessionID"
)

type manager struct{}

func NewContextManager() *manager {
	return &manager{}
}

func (w *manager) InjectSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, keyNameSessionID, sessionID)
}

func (w *manager) ExtractSessionID(ctx context.Context) string {
	rawData := ctx.Value(keyNameSessionID)
	if rawData == nil {
		return ""
	}
	return rawData.(string)
}
