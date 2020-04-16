// +build go1.7,!go1.8

package http_retry

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
)

// Wrap the Body of the Request so it can be read repeatedly in case of retrying
func getBody(r *http.Request) func() (io.ReadCloser, error) {
	if r.Body != nil {
		// Optimise for io.ReadSeeker (e.g file readers) for uploading large files.
		if rs, ok := r.Body.(io.ReadSeeker); ok {
			currentPosition, err := rs.Seek(0, io.SeekCurrent)
			if err != nil {
				return nil
			}
			return func() (closer io.ReadCloser, err error) {
				rs.Seek(currentPosition, io.SeekStart)
				return ioutil.NopCloser(rs), nil
			}
		}
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			return func() (io.ReadCloser, error) {
				return nil, err
			}
		}

		b := bytes.NewBuffer(body)
		r.Body = ioutil.NopCloser(b)

		return func() (io.ReadCloser, error) {
			b := bytes.NewBuffer(body)
			return ioutil.NopCloser(b), nil
		}
	} else {
		// No buffering required as there is no body
		return func() (io.ReadCloser, error) {
			return nil, nil
		}
	}
}
