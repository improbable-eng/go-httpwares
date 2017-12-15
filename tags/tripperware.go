package http_ctxtags

import (
	"net/http"

	"github.com/improbable-eng/go-httpwares"
)

const (
	// TagForCallService is a string naming the ctxtag identifying a "service" grouping for an http.Request (e.g. "github")
	TagForCallService = "http.call.service"
)

// Tripperware returns a new client-side ware that injects tags about the request.
func Tripperware(opts ...Option) httpwares.Tripperware {
	o := evaluateOptions(opts)
	return func(next http.RoundTripper) http.RoundTripper {
		return httpwares.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			t := Extract(req.Context()) // will allocate a new one if it didn't exist.
			defaultRequestTags(t, req)
			for _, extractor := range o.tagExtractors {
				if output := extractor(req); output != nil {
					for k, v := range output {
						t.Set(k, v)
					}

				}
			}
			if !t.Has(TagForCallService) {
				if svc := o.serviceName; svc != "" {
					t.Set(TagForCallService, svc)
				} else {
					svc := o.serviceNameDetectorFunc(req)
					t.Set(TagForCallService, svc)
				}
			}

			newReq := req.WithContext(setInContext(req.Context(), t))
			return next.RoundTrip(newReq)
		})
	}
}
