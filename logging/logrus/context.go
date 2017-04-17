// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_logrus

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/mwitkow/go-httpwares/tags"
	"golang.org/x/net/context"
)

type ctxMarker struct{}

var (
	ctxMarkerKey = &ctxMarker{}
)

// Extract takes the call-scoped logrus.Entry from grpc_logrus middleware.
//
// The logger will have fields pre-populated using http_ctxtags.
//
// If the http_logrus middleware wasn't used, a no-op `logrus.Entry` is returned. This makes it safe to use regardless.
func Extract(req *http.Request) *logrus.Entry {
	return ExtractFromContext(req.Context())
}

// Extract takes the call-scoped logrus.Entry from grpc_logrus middleware.
//
// The logger will have fields pre-populated using http_ctxtags.
//
// If the http_logrus middleware wasn't used, a no-op `logrus.Entry` is returned. This makes it safe to use regardless.
func ExtractFromContext(ctx context.Context) *logrus.Entry {
	l, ok := ctx.Value(ctxMarkerKey).(*logrus.Entry)
	if !ok {
		return logrus.NewEntry(nullLogger)
	}
	// Add grpc_ctxtags tags metadata until now.
	return l.WithFields(logrus.Fields(http_ctxtags.ExtractInboundFromCtx(ctx).Values()))
}

func toContext(ctx context.Context, entry *logrus.Entry) context.Context {
	return context.WithValue(ctx, ctxMarkerKey, entry)

}
