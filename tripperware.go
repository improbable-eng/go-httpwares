package httpwares

import "net/http"

// RoundTripperFunc wraps a func to make it into a http.RoundTripper. Similar to http.HandleFunc.
type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// Tripperware is a signature for all http client-side middleware.
type Tripperware func(http.RoundTripper) http.RoundTripper

// TrpperwareChain is a chain of tripperware before dispatching.
type TripperwareChain []Tripperware

// Forge takes a chain and finalizes it, attaching it to a final RoundTripper.
func (chain TripperwareChain) Forge(final http.RoundTripper) http.RoundTripper {
	next := final
	for i := len(chain) - 1; i >= 0; i-- {
		next = chain[i](next)
	}
	return next
}
