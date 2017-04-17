// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_ctxtags

import (
	"context"
	"net/http"
)

type ctxMarker struct{}

var (
	// serversideMarker is the Context value marker used by *all* server-side middleware.
	serversideMarker = &ctxMarker{}

	// clientsideMarker is the Context value marker used by *all* client-side tripperware.
	clientsideMarker = &ctxMarker{}
)

// Tags is the struct used for storing request tags between Context calls.
// This object is *not* thread safe, and should be handled only in the context of the request.
type Tags struct {
	values map[string]interface{}
}

// Set sets the given key in the metadata tags.
func (t *Tags) Set(key string, value interface{}) *Tags {
	t.values[key] = value
	return t
}

// Has checks if the given key exists.
func (t *Tags) Has(key string) bool {
	_, ok := t.values[key]
	return ok
}

// Values returns a map of key to values.
// Do not modify the underlying map, please use Set instead.
func (t *Tags) Values() map[string]interface{} {
	return t.values
}

// ExtractInbound returns a pre-existing Tags object in the request's Context meant for server-side.
// If the context wasn't set in the Middleware, a no-op Tag storage is returned that will *not* be propagated in context.
func ExtractInbound(req *http.Request) *Tags {
	return ExtractInboundFromCtx(req.Context())
}

// ExtractInbounfFromCtx returns a pre-existing Tags object in the request's Context.
// If the context wasn't set in a tag interceptor, a no-op Tag storage is returned that will *not* be propagated in context.
func ExtractInboundFromCtx(ctx context.Context) *Tags {
	t, ok := ctx.Value(serversideMarker).(*Tags)
	if !ok {
		return &Tags{values: make(map[string]interface{})}
	}
	return t
}

func setInboundInContext(ctx context.Context, tags *Tags) context.Context {
	return context.WithValue(ctx, clientsideMarker, tags)
}

// ExtractOutbound returns a pre-existing Tags object in the request's Context meant for server-side.
// If the context wasn't set in the Middleware, a no-op Tag storage is returned that will *not* be propagated in context.
func ExtractOutbound(req *http.Request) *Tags {
	return ExtractOutboundFromCtx(req.Context())
}

// ExtractInbounfFromCtx returns a pre-existing Tags object in the request's Context.
// If the context wasn't set in a tag interceptor, a no-op Tag storage is returned that will *not* be propagated in context.
func ExtractOutboundFromCtx(ctx context.Context) *Tags {
	t, ok := ctx.Value(clientsideMarker).(*Tags)
	if !ok {
		return &Tags{values: make(map[string]interface{})}
	}
	return t
}

func setOutboundInCtx(ctx context.Context, tags *Tags) context.Context {
	t, ok := ctx.Value(clientsideMarker).(*Tags)
	if ok && t == tags { // points to same variable, no point setting.
		return ctx
	}
	return context.WithValue(ctx, clientsideMarker, tags)
}

func setOutboundInRequest(req *http.Request, tags *Tags) *http.Request {
	t, ok := req.Context().Value(clientsideMarker).(*Tags)
	if ok && t == tags { // points to same variable, no point setting.
		return req
	}
	return req.WithContext(context.WithValue(req.Context(), clientsideMarker, tags))
}
