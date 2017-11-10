// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_ctxtags

import (
	"net"
	"net/http"
	"strings"

	"github.com/improbable-eng/go-httpwares"
)

// Middleware returns a http.Handler middleware values for request tags.
//
// handlerGroupName specifies a logical name for a group of handlers.
func Middleware(handlerGroupName string, opts ...Option) httpwares.Middleware {
	o := evaluateOptions(opts)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			t := ExtractInboundFromCtx(req.Context()) // will allocate a new one if it didn't exist.
			needUpdatingContext := true
			if len(t.Values()) > 0 {
				needUpdatingContext = false
			}
			defaultRequestTags(t, req)
			for _, extractor := range o.tagExtractors {
				if output := extractor(req); output != nil {
					for k, v := range output {
						t.Set(k, v)
					}
				}
			}
			t.Set(TagForHandlerGroup, handlerGroupName)
			newReq := req
			if needUpdatingContext {
				newReq = req.WithContext(setInboundInContext(req.Context(), t))
			}
			next.ServeHTTP(resp, newReq)
		})
	}
}

// HandlerName is a piece of middleware that is meant to be used right around an htpt.Handler that will tag it with a
// given service name and method name.
//
// This tag will be used for tracing, logging and monitoring purposes. This *needs* to be set in a chain of
// Middleware that has `http_ctxtags.Middleware` before it.
func HandlerName(handlerName string) httpwares.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			t := ExtractInbound(req)
			t.Set(TagForHandlerName, handlerName)
			next.ServeHTTP(resp, req)
		})
	}
}

func defaultRequestTags(t *Tags, req *http.Request) {
	if addr := req.RemoteAddr; addr != "" {
		if strings.Contains(addr, ":") {
			if host, port, err := net.SplitHostPort(addr); err == nil {
				t.Set("peer.address", host)
				t.Set("peer.port", port)
			}
		} else {
			t.Set("peer.address", addr)
		}
	}
	host := req.URL.Host
	if host == "" {
		host = req.Host
	}
	t.Set("http.host", host)
}
