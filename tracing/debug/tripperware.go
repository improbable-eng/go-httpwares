// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_debug

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/improbable-eng/go-httpwares"
	"github.com/improbable-eng/go-httpwares/tags"
	"golang.org/x/net/trace"
)

const (
	headerMaxLength = 100
)

// Tripperware returns a piece of client-side Tripperware that puts requests on the `/debug/requests` page.
//
// The data logged will be: request headers, request ctxtags, response headers and response length.
func Tripperware(opts ...Option) httpwares.Tripperware {
	o := evaluateOptions(opts)
	return func(next http.RoundTripper) http.RoundTripper {
		return httpwares.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			if o.filterFunc != nil && !o.filterFunc(req) {
				return next.RoundTrip(req)

			}
			tr := trace.New(operationNameFromUrl(req), req.URL.String())
			defer tr.Finish()

			tr.LazyPrintf("%v %v HTTP/%d.%d", req.Method, req.URL, req.ProtoMajor, req.ProtoMinor)
			tr.LazyPrintf("%s", fmtHeaders(req.Header))

			resp, err := next.RoundTrip(req)

			tr.LazyPrintf("%s", fmtTags(http_ctxtags.ExtractInbound(req).Values()))

			if err != nil {
				tr.LazyPrintf("Error on response: %v", err)
				tr.SetError()
			} else {
				tr.LazyPrintf("HTTP/%d.%d %d %s", resp.ProtoMajor, resp.ProtoMinor, resp.StatusCode, resp.Status)
				tr.LazyPrintf("%s", fmtHeaders(resp.Header))
				if o.statusCodeErrorFunc(resp.StatusCode) {
					tr.SetError()
				}
			}
			return resp, err
		})
	}
}

func operationNameFromUrl(req *http.Request) string {
	if tags := http_ctxtags.ExtractOutbound(req); tags.Has(http_ctxtags.TagForCallService) {
		vals := tags.Values()
		return fmt.Sprintf("%v.%s", vals[http_ctxtags.TagForCallService], req.Method)
	}
	return fmt.Sprintf("%s%s", req.URL.Host, req.URL.Path)
}

func fmtTags(t map[string]interface{}) *bytes.Buffer {
	var b bytes.Buffer
	b.WriteString("tags:")
	for k, v := range t {
		fmt.Fprintf(&b, " %v=%q", k, v)
	}
	return &b
}

func fmtHeaders(h http.Header) *bytes.Buffer {
	var buf bytes.Buffer
	for k := range h {
		v := h.Get(k)
		l := buf.Len()
		if len(k) > headerMaxLength {
			k = k[:headerMaxLength]
		}
		if len(v) > headerMaxLength {
			v = v[:headerMaxLength]
		}
		fmt.Fprintf(&buf, "%v: %v", k, v)
		if buf.Len() > l+headerMaxLength {
			buf.Truncate(l + headerMaxLength)
			fmt.Fprint(&buf, " (header truncated)")
		}
		buf.WriteByte('\n')
	}
	if buf.Len() > 0 {
		buf.Truncate(buf.Len() - 1)
	}
	return &buf
}
