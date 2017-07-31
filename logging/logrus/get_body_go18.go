// +build go1.8

package http_logrus

import (
	"net/http"
	"io"
)

func getBody(r *http.Request) func()(io.ReadCloser, error) {
	return r.GetBody
}