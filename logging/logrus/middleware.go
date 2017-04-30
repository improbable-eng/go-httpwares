// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_logrus

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"strings"
	"time"

	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/mwitkow/go-httpwares"
)

var (
	// SystemField is used in every log statement made through http_logrus. Can be overwritten before any initialization code.
	SystemField = "http"
)

// Middleware is a server-side http ware for logging using logrus.
//
// All handlers will have a Logrus logger in their context, which can be fetched using `http_logrus.Extract`.
func Middleware(entry *logrus.Entry, opts ...Option) httpwares.Middleware {
	return func(nextHandler http.Handler) http.Handler {
		o := evaluateMiddlewareOpts(opts)
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			wrappedResp := httpwares.WrapResponseWriter(resp)
			newEntry := entry.WithFields(
				logrus.Fields{
					"system":                    SystemField,
					"span.kind":                 "server",
					"http.url.path":             req.URL.Path,
					"http.proto_major":          req.ProtoMajor,
					"http.request.length_bytes": req.ContentLength,
				})
			newReq := req.WithContext(toContext(req.Context(), newEntry))
			if o.requestCaptureFunc(req) {
				if err := captureMiddlewareRequestContent(req, Extract(newReq)); err != nil {
					// this is *really* bad, we failed to read a body because of a read error.
					resp.WriteHeader(500)
					Extract(newReq).WithError(err).Warningf("error in logrus middleware on body read")
					return
				}
			}
			var capture *responseCapture
			wrappedResp.ObserveWriteHeader(func(w httpwares.WrappedResponseWriter, code int) {
				if o.responseCaptureFunc(req, code) {
					capture = captureMiddlewareResponseContent(w, Extract(newReq))
				}
			})
			startTime := time.Now()
			nextHandler.ServeHTTP(wrappedResp, newReq)
			capture.finish() // captureResponse has a nil check, this can be nil

			postCallFields := logrus.Fields{
				"http.status":  wrappedResp.StatusCode(),
				"http.time_ms": timeDiffToMilliseconds(startTime),
			}
			level := o.levelFunc(wrappedResp.StatusCode())
			levelLogf(
				Extract(newReq).WithFields(postCallFields), // re-extract logger from newCtx, as it may have extra fields that changed in the holder.
				level,
				"handled")
		})
	}
}

func levelLogf(entry *logrus.Entry, level logrus.Level, format string, args ...interface{}) {
	switch level {
	case logrus.DebugLevel:
		entry.Debugf(format, args...)
	case logrus.InfoLevel:
		entry.Infof(format, args...)
	case logrus.WarnLevel:
		entry.Warningf(format, args...)
	case logrus.ErrorLevel:
		entry.Errorf(format, args...)
	case logrus.FatalLevel:
		entry.Fatalf(format, args...)
	case logrus.PanicLevel:
		entry.Panicf(format, args...)
	default:
		// Unexpected logrus value.
		entry.Panicf(format, args...)
	}
}

func timeDiffToMilliseconds(then time.Time) float32 {
	sub := time.Now().Sub(then).Nanoseconds()
	if sub < 0 {
		return 0.0
	}
	return float32(sub/1000) / 1000.0
}

func captureMiddlewareRequestContent(req *http.Request, entry *logrus.Entry) error {
	if req.ContentLength <= 0 || req.Body == nil {
		// -1 value means that the length cannot be determined, and that it is probably a multipart stremaing call
		if req.ContentLength != 0 || req.Body == nil {
			entry.Infof("request body capture skipped, content length negative")
		}
		return nil
	}
	content, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	// Make sure we give the Response back its body so the client can read it.
	req.Body = ioutil.NopCloser(bytes.NewReader(content))
	if strings.HasPrefix(strings.ToLower(req.Header.Get("content-type")), "application/json") {
		entry.WithField("http.request.body_json", json.RawMessage(content)).Info("request body captured in http.request.body_json field")
	} else {
		entry.WithField("http.request.body_raw", base64.StdEncoding.EncodeToString(content)).Info("request body captured in http.request.body_raw field")
	}
	return nil
}

type responseCapture struct {
	content bytes.Buffer
	isJson  bool
	entry   *logrus.Entry
}

func (c *responseCapture) observeWrite(resp httpwares.WrappedResponseWriter, buf []byte, n int, err error) {
	if err == nil {
		c.content.Write(buf[:n])
	}
}

func (c *responseCapture) finish() {
	if c == nil {
		return
	}
	if c.content.Len() == 0 {
		return
	}
	if c.isJson {
		e := c.entry.WithField("http.response.body_json", json.RawMessage(c.content.Bytes()))
		e.Info("response body captured in http.response.body_json field")
	} else {
		e := c.entry.WithField("http.response.body_raw", base64.StdEncoding.EncodeToString(c.content.Bytes()))
		e.Info("response body captured in http.response.body_raw field")
	}
}

func captureMiddlewareResponseContent(w httpwares.WrappedResponseWriter, entry *logrus.Entry) *responseCapture {
	c := &responseCapture{entry: entry}
	if te := w.Header().Get("transfer-encoding"); te != "" {
		entry.Infof("response body capture skipped, transfer encoding is not identity")
		return nil
	}
	c.isJson = headerIsJson(w.Header())
	w.ObserveWrite(c.observeWrite)
	return c
}
