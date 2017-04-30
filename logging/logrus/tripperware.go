// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_logrus

import (
	"net/http"
	"time"

	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/mwitkow/go-httpwares"
	"github.com/mwitkow/go-httpwares/tags"
)

// Tripperware is a server-side http ware for logging using logrus.
//
// This tripperware *does not* propagate a context-based logger, but act as a logger of requests.
// This includes logging of errors.
func Tripperware(entry *logrus.Entry, opts ...Option) httpwares.Tripperware {
	return func(next http.RoundTripper) http.RoundTripper {
		o := evaluateTripperwareOpts(opts)
		return httpwares.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			startTime := time.Now()
			fields := logrus.Fields{
				"system":                    SystemField,
				"span.kind":                 "client",
				"http.url.path":             req.URL.Path,
				"http.request.length_bytes": req.ContentLength,
			}
			for k, v := range http_ctxtags.ExtractOutbound(req).Values() {
				fields[k] = v
			}
			if o.requestCaptureFunc(req) {
				if err := captureTripperwareRequestContent(req, entry.WithFields(fields)); err != nil {
					logError(o, entry.WithFields(fields), err)
					return nil, err // errors reading GetBody and other problems on client side
				}
			}
			resp, err := next.RoundTrip(req)
			fields["http.time_ms"] = timeDiffToMilliseconds(startTime)
			if err != nil {
				logError(o, entry.WithFields(fields), err)
				return nil, err
			}
			fields["http.proto_major"] = resp.ProtoMajor
			fields["http.response.length_bytes"] = resp.ContentLength
			fields["http.status"] = resp.StatusCode
			if o.responseCaptureFunc(req, resp.StatusCode) {
				if err := captureTripperwareResponseContent(resp, entry.WithFields(fields)); err != nil {
					logError(o, entry.WithFields(fields), err)
					return nil, err
				}
			}
			levelLogf(entry.WithFields(fields), o.levelFunc(resp.StatusCode), "request completed")
			return resp, nil
		})
	}
}

func logError(o *options, e *logrus.Entry, err error) {
	levelLogf(e, o.levelForConnectivityError, "request failed to execute, see err")
}

func headerIsJson(header http.Header) bool {
	return strings.HasPrefix(strings.ToLower(header.Get("content-type")), "application/json")
}

func captureTripperwareRequestContent(req *http.Request, entry *logrus.Entry) error {
	// All requests created with http.NewRequest will have a GetBody method set, even if the user created
	// a body manually.
	if req.GetBody == nil {
		if req.Body != nil {
			entry.Infof("request body capture skipped, missing GetBody method while Body set")
		}
		return nil
	}
	bodyReader, err := req.GetBody()
	if err != nil {
		return err
	}
	content, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return err
	}
	if headerIsJson(req.Header) {
		entry.WithField("http.request.body_json", json.RawMessage(content)).Info("request body captured in http.request.body_json field")
	} else {
		entry.WithField("http.request.body_raw", base64.StdEncoding.EncodeToString(content)).Info("request body captured in http.request.body_raw field")
	}
	return nil
}

func captureTripperwareResponseContent(resp *http.Response, entry *logrus.Entry) error {
	if resp.ContentLength <= 0 {
		// TODO(mwitkow): Deal with response.Uncompressed and gzip encoding (Content Length -1).
		if resp.ContentLength != 0 {
			entry.Infof("response body capture skipped, content length negative")
		}
		return nil
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err // this is an error form the response reading, potentially a connection failure
	}
	// Make sure we give the Response back its body so the client can read it.
	resp.Body = ioutil.NopCloser(bytes.NewReader(content))
	if headerIsJson(resp.Header) {
		entry.WithField("http.response.body_json", json.RawMessage(content)).Info("request body captured in http.response.body_json field")
	} else {
		entry.WithField("http.response.body_raw", base64.StdEncoding.EncodeToString(content)).Info("request body captured in http.response.body_raw field")
	}
	return nil
}
