package logger

import (
	"context"

	"github.com/mkhsnw/golang-starter-kit/internal/foundation/appcontext" // package appcontext
	"github.com/sirupsen/logrus"
)

// WithContext extracts the RequestContext from the standard context and returns a logrus.Entry with the fields.
func WithContext(ctx context.Context, log *logrus.Logger) *logrus.Entry {
	if ctx == nil {
		return logrus.NewEntry(log)
	}

	reqCtx := appcontext.GetRequestContext(ctx)
	if reqCtx == nil {
		return logrus.NewEntry(log)
	}

	return log.WithFields(logrus.Fields{
		"request_id": reqCtx.RequestID,
		"trace_id":   reqCtx.TraceID,
		"user_id":    reqCtx.UserID,
		"ip":         reqCtx.IPAddress,
	})
}
