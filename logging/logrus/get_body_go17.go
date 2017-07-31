// +build go1.7,!go1.8

package http_logrus

import (
	"io"
	"net/http"
	"reflect"
)

func getBody(r *http.Request) func() (io.ReadCloser, error) {
	if r.Body != nil {
		var clonedBody io.ReadCloser
		clonedBody = reflect.New(reflect.ValueOf(r.Body).Elem().Type()).Interface().(io.ReadCloser)
		return func() (io.ReadCloser, error) {
			return clonedBody, nil
		}
	}
	return nil
}
