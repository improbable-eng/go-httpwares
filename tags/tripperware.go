package http_ctxtags

import (
	"net/http"

	"github.com/mwitkow/go-httpwares"
)

// Tripperware returns a new client-side ware that injects tags about the request.
func Tripperware(opts ...Option) httpwares.Tripperware {
	o := evaluateOptions(opts)
	return func(next http.RoundTripper) http.RoundTripper {
		return httpwares.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			t := ExtractOutbound(req) // will allocate a new one if it didn't exist.
			defaultRequestTags(t, req)
			for _, extractor := range o.tagExtractors {
				if output := extractor(req); output != nil {
					for k, v := range output {
						t.Set(k, v)
					}

				}
			}
			if !t.Has(TagForCallService) {
				t.Set(TagForCallService, o.defaultServiceName)
			}

			newReq := setOutboundInRequest(req, t)
			return next.RoundTrip(newReq)
		})
	}
}
