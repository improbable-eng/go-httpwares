// +build go1.7,!go1.8

package http_retry_test

import "net/http"

func removeGetBody(r *http.Request) {
	// no-op
}
