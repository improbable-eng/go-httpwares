// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_ctxtags

import (
	"net/http"

	"github.com/pressly/chi"
)

// ChiRouteTagExtractor extracts chi router information and puts them into tags.
func ChiRouteTagExtractor(req *http.Request) map[string]interface{} {
	if routeCtx, ok := req.Context().Value(chi.RouteCtxKey).(*chi.Context); ok {
		val := map[string]interface{}{
			"http.route": routeCtx.RoutePattern,
		}
		for _, param := range routeCtx.URLParams {
			val["http.request.pathparam."+param.Key] = param.Value
		}
		return val
	}
	return nil
}
