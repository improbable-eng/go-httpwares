// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_logrus

import (
	"time"

	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/mwitkow/go-httpwares"
	"golang.org/x/net/context"
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
			nCtx := newContextLogger(req.Context(), entry, req)
			startTime := time.Now()
			nextHandler.ServeHTTP(wrappedResp, req.WithContext(nCtx))
			postCallFields := logrus.Fields{
				"http.status":  wrappedResp.StatusCode(),
				"http.time_ms": timeDiffToMilliseconds(startTime),
			}
			level := o.levelFunc(wrappedResp.StatusCode())
			levelLogf(
				ExtractFromContext(nCtx).WithFields(postCallFields), // re-extract logger from newCtx, as it may have extra fields that changed in the holder.
				level,
				"handled")
		})
	}
}

func newContextLogger(ctx context.Context, entry *logrus.Entry, r *http.Request) context.Context {
	callLog := entry.WithFields(
		logrus.Fields{
			"system":                    SystemField,
			"span.kind":                 "server",
			"http.url.path":             r.URL.Path,
			"http.proto_major":          r.ProtoMajor,
			"http.request.length_bytes": r.ContentLength,
		})
	return toContext(ctx, callLog)
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
