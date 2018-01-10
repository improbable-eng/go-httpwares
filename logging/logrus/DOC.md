# http_logrus
`import "github.com/improbable-eng/go-httpwares/logging/logrus"`

* [Overview](#pkg-overview)
* [Imported Packages](#pkg-imports)
* [Index](#pkg-index)
* [Examples](#pkg-examples)

## <a name="pkg-overview">Overview</a>
`http_logrus` is a HTTP logging middleware for the Logrus logging stack.

It provides both middleware (server-side) and tripperware (client-side) for logging HTTP requests using a user-provided
`logrus.Entry`.

### Middleware server-side logging
The middleware also embeds a request-field scoped `logrus.Entry` (with fields from `ctxlogrus`) inside the `context.Context`
of the `http.Request` that is passed to the executing `http.Handler`. That `logrus.Entry` can be easily extracted using
It accepts a user-configured `logrus.Entry` that will be used for logging completed HTTP calls. The same
`logrus.Entry` will be used for logging completed gRPC calls, and be populated into the `context.Context` passed into
HTTP handler code. To do that, use the `Extract` method (see example below).

The middlewarerequest will be logged at a level indicated by `WithLevels` options, and an example JSON-formatted
log message will look like:

	{
	"@timestamp:" "2006-01-02T15:04:05Z07:00",
	"@level": "info",
	"my_custom.my_string": 1337,
	"custom_tags.string": "something",
	"http.handler.group": "my_service",
	"http.host": "something.local",
	"http.proto_major": 1,
	"http.request.length_bytes": 0,
	"http.status": 201,
	"http.time_ms": 0.095,
	"http.url.path": "/someurl",
	"msg": "handled",
	"peer.address": "127.0.0.1",
	"peer.port": "59141",
	"span.kind": "server",
	"system": "http"
	}

### Tripperware client-side logging
The tripperware uses any `ctxlogrus` to create a request-field scoped `logrus.Entry`. The key one is the `http.call.service`
which by default is auto-detected from the domain but can be overwritten by the `ctxlogrus` initialization.

Most requests and responses won't be loged. By default only client-side connectivity  and 5** responses cause
the outbound requests to be logged, but that can be customized using `WithLevels` and `WithConnectivityError` options. A
typical log message for client side will look like:

	{
	"@timestamp:" "2006-01-02T15:04:05Z07:00",
	"@level": "debug",
	"http.call.service": "googleapis",
	"http.host": "calendar.googleapis.com",
	"http.proto_major": 1,
	"http.request.length_bytes": 0,
	"http.response.length_bytes": 176,
	"http.status": 201,
	"http.time_ms": 4.654,
	"http.url.path": "/someurl",
	"msg": "request completed",
	"span.kind": "client",
	"system": "http"
	}

You can use `Extract` to log into a request-scoped `logrus.Entry` instance in your handler code.
Additional tags to the logger can be added using `ctxlogrus`.

### HTTP Library logging
The `http.Server` takes a logger command. You can use the `AsHttpLogger` to take a user-scoped `logrus.Entry` and log
connectivity or low-level HTTP errors (e.g. TLS handshake problems, badly formed requests etc).

Please see examples and tests for examples of use.

## <a name="pkg-imports">Imported Packages</a>

- [github.com/improbable-eng/go-httpwares](./../..)
- [github.com/improbable-eng/go-httpwares/logging](./..)
- [github.com/improbable-eng/go-httpwares/logging/logrus/ctxlogrus](./ctxlogrus)
- [github.com/sirupsen/logrus](https://godoc.org/github.com/sirupsen/logrus)

## <a name="pkg-index">Index</a>
* [Variables](#pkg-variables)
* [func AsHttpLogger(logger \*logrus.Entry) \*log.Logger](#AsHttpLogger)
* [func ContentCaptureMiddleware(entry \*logrus.Entry, decider http\_logging.ContentCaptureDeciderFunc) httpwares.Middleware](#ContentCaptureMiddleware)
* [func ContentCaptureTripperware(entry \*logrus.Entry, decider http\_logging.ContentCaptureDeciderFunc) httpwares.Tripperware](#ContentCaptureTripperware)
* [func DefaultMiddlewareCodeToLevel(httpStatusCode int) logrus.Level](#DefaultMiddlewareCodeToLevel)
* [func DefaultTripperwareCodeToLevel(httpStatusCode int) logrus.Level](#DefaultTripperwareCodeToLevel)
* [func Middleware(entry \*logrus.Entry, opts ...Option) httpwares.Middleware](#Middleware)
* [func Tripperware(entry \*logrus.Entry, opts ...Option) httpwares.Tripperware](#Tripperware)
* [type CodeToLevel](#CodeToLevel)
* [type Decider](#Decider)
* [type Option](#Option)
  * [func WithConnectivityErrorLevel(level logrus.Level) Option](#WithConnectivityErrorLevel)
  * [func WithDecider(f Decider) Option](#WithDecider)
  * [func WithLevels(f CodeToLevel) Option](#WithLevels)
  * [func WithRequestBodyCapture(deciderFunc func(r \*http.Request) bool) Option](#WithRequestBodyCapture)
  * [func WithRequestFieldExtractor(f RequestFieldExtractorFunc) Option](#WithRequestFieldExtractor)
  * [func WithResponseBodyCapture(deciderFunc func(r \*http.Request, status int) bool) Option](#WithResponseBodyCapture)
  * [func WithResponseFieldExtractor(f ResponseFieldExtractorFunc) Option](#WithResponseFieldExtractor)
* [type RequestFieldExtractorFunc](#RequestFieldExtractorFunc)
* [type ResponseFieldExtractorFunc](#ResponseFieldExtractorFunc)

#### <a name="pkg-examples">Examples</a>
* [Middleware](#example_Middleware)
* [WithRequestFieldExtractor](#example_WithRequestFieldExtractor)
* [WithResponseFieldExtractor](#example_WithResponseFieldExtractor)

#### <a name="pkg-files">Package files</a>
[capture_middleware.go](./capture_middleware.go) [capture_tripperware.go](./capture_tripperware.go) [doc.go](./doc.go) [get_body_go18.go](./get_body_go18.go) [httplogger.go](./httplogger.go) [middleware.go](./middleware.go) [options.go](./options.go) [tripperware.go](./tripperware.go) 

## <a name="pkg-variables">Variables</a>
``` go
var (
    // SystemField is used in every log statement made through http_logrus. Can be overwritten before any initialization code.
    SystemField = "http"
)
```

## <a name="AsHttpLogger">func</a> [AsHttpLogger](./httplogger.go#L10)
``` go
func AsHttpLogger(logger *logrus.Entry) *log.Logger
```
AsHttpLogger returns the given logrus instance as an HTTP logger.

## <a name="ContentCaptureMiddleware">func</a> [ContentCaptureMiddleware](./capture_middleware.go#L29)
``` go
func ContentCaptureMiddleware(entry *logrus.Entry, decider http_logging.ContentCaptureDeciderFunc) httpwares.Middleware
```
ContentCaptureMiddleware is a server-side http ware for logging contents of HTTP requests and responses (body and headers).

Only requests with a set Content-Length will be captured, with no streaming or chunk encoding supported.
Only responses with Content-Length set are captured, no gzipped, chunk-encoded responses are supported.

The body will be recorded as a separate log message. Body of `application/json` will be captured as
http.request.body_json (in structured JSON form) and others will be captured as http.request.body_raw logrus field
(raw base64-encoded value).

This *must* be used together with http_logrus.Middleware, as it relies on the logger provided there. However, you can
override the `logrus.Entry` that is used for logging, allowing for logging to a separate backend (e.g. a different file).

## <a name="ContentCaptureTripperware">func</a> [ContentCaptureTripperware](./capture_tripperware.go#L23)
``` go
func ContentCaptureTripperware(entry *logrus.Entry, decider http_logging.ContentCaptureDeciderFunc) httpwares.Tripperware
```
ContentCaptureTripperware is a client-side http ware for logging contents of HTTP requests and responses (body and headers).

Only requests with a set GetBody field will be captured (strings, bytes etc).
Only responses with Content-Length are captured, with no support for chunk-encoded responses.

The body will be recorded as a separate log message. Body of `application/json` will be captured as
http.request.body_json (in structured JSON form) and others will be captured as http.request.body_raw logrus field
(raw base64-encoded value).

## <a name="DefaultMiddlewareCodeToLevel">func</a> [DefaultMiddlewareCodeToLevel](./options.go#L117)
``` go
func DefaultMiddlewareCodeToLevel(httpStatusCode int) logrus.Level
```
DefaultMiddlewareCodeToLevel is the default of a mapper between HTTP server-side status codes and logrus log levels.

## <a name="DefaultTripperwareCodeToLevel">func</a> [DefaultTripperwareCodeToLevel](./options.go#L130)
``` go
func DefaultTripperwareCodeToLevel(httpStatusCode int) logrus.Level
```
DefaultTripperwareCodeToLevel is the default of a mapper between HTTP client-side status codes and logrus log levels.

## <a name="Middleware">func</a> [Middleware](./middleware.go#L23)
``` go
func Middleware(entry *logrus.Entry, opts ...Option) httpwares.Middleware
```
Middleware is a server-side http ware for logging using logrus.

All handlers will have a Logrus logger in their context, which can be fetched using `ctxlogrus.Extract`.

#### Example:

<details>
<summary>Click to expand code.</summary>

```go
h := http.NewServeMux()
h.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    ctxlogrus.Extract(r.Context()).Info("logging")
})
hm := Middleware(logrus.WithField("test", "abc"))(h)
if err := http.ListenAndServe(":8080", hm); err != nil {
    panic(err)
}
```

</details>

## <a name="Tripperware">func</a> [Tripperware](./tripperware.go#L15)
``` go
func Tripperware(entry *logrus.Entry, opts ...Option) httpwares.Tripperware
```
Tripperware is a server-side http ware for logging using logrus.

This tripperware *does not* propagate a context-based logger, but act as a logger of requests.
This includes logging of errors.

## <a name="CodeToLevel">type</a> [CodeToLevel](./options.go#L56)
``` go
type CodeToLevel func(httpStatusCode int) logrus.Level
```
CodeToLevel user functions define the mapping between HTTP status codes and logrus log levels.

## <a name="Decider">type</a> [Decider](./options.go#L114)
``` go
type Decider func(w httpwares.WrappedResponseWriter, r *http.Request) bool
```
Decider function defines rules for suppressing any interceptor logs

## <a name="Option">type</a> [Option](./options.go#L53)
``` go
type Option func(*options)
```

### <a name="WithConnectivityErrorLevel">func</a> [WithConnectivityErrorLevel](./options.go#L69)
``` go
func WithConnectivityErrorLevel(level logrus.Level) Option
```
WithConnectivityErrorLevel customizes

### <a name="WithDecider">func</a> [WithDecider](./options.go#L107)
``` go
func WithDecider(f Decider) Option
```
WithDecider customizes the function for deciding if the middleware logs at the end of the request.

### <a name="WithLevels">func</a> [WithLevels](./options.go#L62)
``` go
func WithLevels(f CodeToLevel) Option
```
WithLevels customizes the function that maps HTTP client or server side status codes to log levels.

By default `DefaultMiddlewareCodeToLevel` is used for server-side middleware, and `DefaultTripperwareCodeToLevel`
is used for client-side tripperware.

### <a name="WithRequestBodyCapture">func</a> [WithRequestBodyCapture](./options.go#L87)
``` go
func WithRequestBodyCapture(deciderFunc func(r *http.Request) bool) Option
```
WithRequestBodyCapture enables recording of request body pre-handling/pre-call.

The body will be recorded as a separate log message. Body of `application/json` will be captured as
http.request.body_json (in structured JSON form) and others will be captured as http.request.body_raw logrus field
(raw base64-encoded value).

For tripperware, only requests with Body of type `bytes.Buffer`, `strings.Reader`, `bytes.Buffer`, or with
a specified `GetBody` function will be captured.

For middleware, only requests with a set Content-Length will be captured, with no streaming or chunk encoding supported.

This option creates a copy of the body per request, so please use with care.

### <a name="WithRequestFieldExtractor">func</a> [WithRequestFieldExtractor](./options.go#L141)
``` go
func WithRequestFieldExtractor(f RequestFieldExtractorFunc) Option
```
WithRequestFieldExtractor adds a field, allowing you to customize what fields get populated from the request.

#### Example:

<details>
<summary>Click to expand code.</summary>

```go
Middleware(logrus.WithField("foo", "bar"),
    WithRequestFieldExtractor(func(req *http.Request) map[string]interface{} {
        return map[string]interface{}{
            "http.request.customFieldA": req.Header.Get("x-custom-header"),
            "http.request.customFieldB": req.Header.Get("x-another-custom-header"),
        }
    }),
)
```

</details>
### <a name="WithResponseBodyCapture">func</a> [WithResponseBodyCapture](./options.go#L100)
``` go
func WithResponseBodyCapture(deciderFunc func(r *http.Request, status int) bool) Option
```
WithResponseBodyCapture enables recording of response body post-handling/post-call.

The body will be recorded as a separate log message. Body of `application/json` will be captured as
http.response.body_json (in structured JSON form) and others will be captured as http.response.body_raw logrus field
(raw base64-encoded value).

Only responses with Content-Length will be captured, with non-default Transfer-Encoding not being supported.

### <a name="WithResponseFieldExtractor">func</a> [WithResponseFieldExtractor](./options.go#L148)
``` go
func WithResponseFieldExtractor(f ResponseFieldExtractorFunc) Option
```
WithRequestFieldExtractor adds a field, allowing you to customize what fields get populated from the response.

#### Example:

<details>
<summary>Click to expand code.</summary>

```go
Middleware(logrus.WithField("foo", "bar"),
    WithResponseFieldExtractor(func(res httpwares.WrappedResponseWriter) map[string]interface{} {
        return map[string]interface{}{
            "http.response.customFieldC": res.StatusCode(),
        }
    }),
)
```

</details>

## <a name="RequestFieldExtractorFunc">type</a> [RequestFieldExtractorFunc](./options.go#L155)
``` go
type RequestFieldExtractorFunc func(req *http.Request) map[string]interface{}
```
RequestFieldExtractorFunc is a signature of user-customizable functions for extracting log fields from requests.

## <a name="ResponseFieldExtractorFunc">type</a> [ResponseFieldExtractorFunc](./options.go#L158)
``` go
type ResponseFieldExtractorFunc func(res httpwares.WrappedResponseWriter) map[string]interface{}
```
ResponseFieldExtractorFunc is a signature of user-customizable functions for extracting log fields from responses.

- - -
Generated by [godoc2ghmd](https://github.com/GandalfUK/godoc2ghmd)