// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_ctxtags

import (
	"net/http"

	"github.com/pressly/chi"
)

// ChiRouteTagExtractor extracts chi router information and puts them into tags.
//
// By default it will treat the route pattern as the handler name.
func ChiRouteTagExtractor(req *http.Request) map[string]interface{} {
	if routeCtx, ok := req.Context().Value(chi.RouteCtxKey).(*chi.Context); ok {
		val := map[string]interface{}{
			TagForHandlerName: routeCtx.RoutePath,
		}
		// TODO(bplotka): Find a way to obtain params from chi routeCtx.URLParams (routeParams struct).
		// Internal keys & values are no longer public, you can only ask for known keys
		// using "func (x *Context) URLParam(key string) string".
		return val
	}
	return nil
}
