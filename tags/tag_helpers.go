package http_ctxtags

import (
	"net/http"
)

const (
	// TagForCallService is a string naming the ctxtag identifying a "service" grouping for an http.Request (e.g. "github")
	TagForCallService = "http.call.service"
	// TagForCallMethod is a string naming the ctxtag identifying a "method" in a "service" for an http.Request (e.g. "login")
	TagForCallMethod = "http.call.method"

	// TagForHandlerService is a string naming the ctxtag identifying a "service" grouping of http.Handlers (e.g. auth).
	TagForHandlerService = "http.handler.service"
	// TagForHandlerMethod is a string naming the ctxtag identifying a logical "method" of the http.Handler (e.g. exchange_token).
	TagForHandlerMethod = "http.handler.method"
)

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
			t.Set(TagForHandlerService, serviceName)
		}
		if methodName != "" {
			t.Set(TagForHandlerMethod, methodName)
		}
		handler.ServeHTTP(resp, req)
	})
}

// TagRequest is a helper that identifies an `http.Request` service name and method.
//
// This is useful to add "service/method" semantics to your external calls. This tag will be used for tracing, logging
// and monitoring purposes. In order for it to work, the invoked `http.Client` needs to have `http_ctxtags.Tripperware`
// in its Roundtripper chain.
//
// You can pass in an empty serviceName, in which case it will inherit it from the `http_ctxtags.Tripperware` configuration.
//
// It returns a new Request object.
func TagRequest(req *http.Request, serviceName string, methodName string) *http.Request {
	t := ExtractInbound(req)
	if serviceName != "" {
		t.Set(TagForCallService, serviceName)
	}
	if methodName != "" {
		t.Set(TagForCallMethod, methodName)
	}
	return setOutboundInRequest(req, t)
}
