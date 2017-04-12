// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package httpwares_ctxtags

import (
	"net"
	"net/http"
	"strings"

	"github.com/mwitkow/go-httpwares"
)

// Middleware returns a http.Handler middleware values for request tags.
func Middleware(opts ...Option) httpwares.Middleware {
	o := evaluateOptions(opts)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			t := ExtractFromContext(req.Context()) // will allocate a new one if it didn't exist.
			defaultRequestTags(t, req)
			for _, extractor := range o.tagExtractors {
				if output := extractor(req); output != nil {
					for k, v := range output {
						t.Set(k, v)
					}
				}
			}
			next.ServeHTTP(resp, req.WithContext(setInContext(req.Context(), t)))
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
