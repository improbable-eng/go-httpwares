// Copyright (c) Improbable Worlds Ltd, All Rights Reserved

package ctx_logrus

import (
	"github.com/improbable-eng/go-httpwares/tags"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type ctxMarker struct{}

var (
	ctxMarkerKey = &ctxMarker{}
)

// Extract takes the call-scoped logrus.Entry from http_ctxtags middleware.
//
// The logger will have fields pre-populated using http_ctxtags.
//
// If the http_logrus middleware wasn't used, a no-op `logrus.Entry` is returned. This makes it safe to use regardless.
func Extract(ctx context.Context) *logrus.Entry {
	l, ok := ctx.Value(ctxMarkerKey).(*logrus.Entry)
	if !ok {
		return logrus.NewEntry(nullLogger)
	}
	// Add http_ctxtags tags metadata until now.
	return l.WithFields(logrus.Fields(http_ctxtags.Extract(ctx).Values()))
}

// ToContext adds the logrus.Entry to the context for extraction later.
// Returning the new context that has been created.
func ToContext(ctx context.Context, entry *logrus.Entry) context.Context {
	return context.WithValue(ctx, ctxMarkerKey, entry)

}
