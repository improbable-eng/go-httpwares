package http_ctxtags

import "net/http"

// TagHandler is a helper wrapper for http.Handler that will tag it with a given service name and method name.
//
// This tag will be used for tracing, logging and monitoring purposes. This *needs* to be set in a chain of
// Middleware that has `http_ctxtags.Middleware` before it.
//
// You can pass in an empty serviceName, in which case it will inherit it from the `http_ctxtags.Middleware` configuration.
func TagHandler(serviceName string, methodName string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		t := ExtractInbound(req)
		if serviceName != "" {
			t.Set("http.service", serviceName)
		}
		if methodName != "" {
			t.Set("http.method", methodName)
		}
		handler.ServeHTTP(resp, req)
	})
}
