// +build go1.8

package http_retry

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
)

func getBody(r *http.Request) func() (io.ReadCloser, error) {
	if r.GetBody != nil {
		return r.GetBody
	} else if r.Body != nil {
		// If the GetBody function does not exist, setup our own buffering for the body to allow re-reads
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return func() (io.ReadCloser, error) {
				return nil, err
			}
		}
		return func() (io.ReadCloser, error) {
			// Create a new buffer containing the body data each time
			return ioutil.NopCloser(bytes.NewBuffer(data)), nil
		}
	} else {
		// No buffering required as there is no body
		return func() (io.ReadCloser, error) {
			return nil, nil
		}
	}
}
