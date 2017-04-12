# http_logrus
--
    import "github.com/mwitkow/go-httpwares/logging/logrus"

`http_logrus` is a HTTP logging middleware for the Logrus logging stack.

It accepts a user-configured `logrus.Entry` that will be used for logging
completed HTTP calls. The same `logrus.Entry` will be used for logging completed
gRPC calls, and be populated into the `context.Context` passed into HTTP handler
code.

You can use `Extract` to log into a request-scoped `logrus.Entry` instance in
your handler code. Additional tags to the logger can be added using
`http_ctxtags`.

Logrus can also be made as a backend for HTTP library internals. For that use
`AsHttpLogger`.

Please see examples and tests for examples of use.

## Usage

```go
var (
	// SystemField is used in every log statement made through http_logrus. Can be overwritten before any initialization code.
	SystemField = "http"
)
```

#### func  AsHttpLogger

```go
func AsHttpLogger(logger *logrus.Entry) *log.Logger
```
AsHttpLogger returns the given logrus instance as an HTTP logger.

#### func  DefaultCodeToLevel

```go
func DefaultCodeToLevel(httpStatusCode int) logrus.Level
```
DefaultCodeToLevel is the default implementation of gRPC return codes and
interceptor log level.

#### func  Extract

```go
func Extract(req *http.Request) *logrus.Entry
```
Extract takes the call-scoped logrus.Entry from grpc_logrus middleware.

The logger will have fields pre-populated using http_ctxtags.

If the http_logrus middleware wasn't used, a no-op `logrus.Entry` is returned.
This makes it safe to use regardless.

#### func  ExtractFromContext

```go
func ExtractFromContext(ctx context.Context) *logrus.Entry
```
Extract takes the call-scoped logrus.Entry from grpc_logrus middleware.

The logger will have fields pre-populated using http_ctxtags.

If the http_logrus middleware wasn't used, a no-op `logrus.Entry` is returned.
This makes it safe to use regardless.

#### func  Middleware

```go
func Middleware(entry *logrus.Entry, opts ...Option) func(http.Handler) http.Handler
```
Middleware is a server-side http ware for logging using logrus.

#### type CodeToLevel

```go
type CodeToLevel func(httpStatusCode int) logrus.Level
```

CodeToLevel function defines the mapping between gRPC return codes and
interceptor log level.

#### type Option

```go
type Option func(*options)
```


#### func  WithLevels

```go
func WithLevels(f CodeToLevel) Option
```
WithLevels customizes the function for mapping gRPC return codes and interceptor
log level statements.
