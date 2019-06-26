// +build go1.7,!go1.8

package http_retry_test

func removeGetBody(r *http.Request) {
	// no-op
}
