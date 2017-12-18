// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_logrus

import (
	"time"

	"net/http"

	"github.com/improbable-eng/go-httpwares"
	"github.com/improbable-eng/go-httpwares/tags/logrus"
	"github.com/sirupsen/logrus"
)

var (
	// SystemField is used in every log statement made through http_logrus. Can be overwritten before any initialization code.
	SystemField = "http"
)

// Middleware is a server-side http ware for logging using logrus.
//
// All handlers will have a Logrus logger in their context, which can be fetched using `ctx_logrus.Extract`.
func Middleware(entry *logrus.Entry, opts ...Option) httpwares.Middleware {
	return func(nextHandler http.Handler) http.Handler {
		o := evaluateMiddlewareOpts(opts)
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			wrappedResp := httpwares.WrapResponseWriter(resp)
			newEntry := entry.WithFields(newServerRequestFields(req))
			newReq := req.WithContext(ctx_logrus.ToContext(req.Context(), newEntry))
			var capture *responseCapture
			wrappedResp.ObserveWriteHeader(func(w httpwares.WrappedResponseWriter, code int) {
				if o.responseCaptureFunc(req, code) {
					capture = captureMiddlewareResponseContent(w, ctx_logrus.Extract(newReq))
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
				ctx_logrus.Extract(newReq).WithFields(postCallFields), // re-extract logger from newCtx, as it may have extra fields that changed in the holder.
				level,
				"handled")
		})
	}
}

func newServerRequestFields(req *http.Request) logrus.Fields {
	return logrus.Fields{
		"system":                    SystemField,
		"span.kind":                 "server",
		"http.url.path":             req.URL.Path,
		"http.proto_major":          req.ProtoMajor,
		"http.request.length_bytes": req.ContentLength,
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
