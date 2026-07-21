package appcontext

import (
	stdcontext "context"
)

// RequestContext holds request-scoped information.
type RequestContext struct {
	RequestID string
	UserID    string
	Role      string
	TraceID   string
	Locale    string
	IPAddress string
	UserAgent string
}

type contextKey string

const requestContextKey = contextKey("requestContext")

// WithRequestContext injects the RequestContext into the standard context.Context.
func WithRequestContext(ctx stdcontext.Context, reqCtx *RequestContext) stdcontext.Context {
	return stdcontext.WithValue(ctx, requestContextKey, reqCtx)
}

// GetRequestContext extracts the RequestContext from the standard context.Context.
// Returns nil if it's not found.
func GetRequestContext(ctx stdcontext.Context) *RequestContext {
	if reqCtx, ok := ctx.Value(requestContextKey).(*RequestContext); ok {
		return reqCtx
	}
	return nil
}

// GetUserID extracts just the UserID if available, useful for quick checks.
func GetUserID(ctx stdcontext.Context) string {
	if reqCtx := GetRequestContext(ctx); reqCtx != nil {
		return reqCtx.UserID
	}
	return ""
}
