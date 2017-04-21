# http_debug
--
    import "github.com/mwitkow/go-httpwares/tracing/debug"


## Usage

#### func  DefaultIsStatusCodeAnError

```go
func DefaultIsStatusCodeAnError(statusCode int) bool
```
DefaultIsStatusCodeAnError defines a function that says whether a given request
is an error based on a code.

#### func  Middleware

```go
func Middleware(opts ...Option) httpwares.Middleware
```
Middleware returns a http.Handler middleware that writes inbound requests to
/debug/request.

The data logged will be: request headers, request ctxtags, response headers and
response length.

#### func  Tripperware

```go
func Tripperware(opts ...Option) httpwares.Tripperware
```
Tripperware returns a piece of client-side Tripperware that puts requests on the
`/debug/requests` page.

The data logged will be: request headers, request ctxtags, response headers and
response length.

#### type FilterFunc

```go
type FilterFunc func(req *http.Request) bool
```

FilterFunc allows users to provide a function that filters out certain methods
from being traced.

If it returns false, the given request will not be traced.

#### type IsStatusCodeAnErrorFunc

```go
type IsStatusCodeAnErrorFunc func(statusCode int) bool
```

IsStatusCodeAnErrorFunc allows the customization of which requests are
considered errors in the tracing system.

#### type Option

```go
type Option func(*options)
```


#### func  WithFilterFunc

```go
func WithFilterFunc(f FilterFunc) Option
```
WithFilterFunc customizes the function used for deciding whether a given call is
traced or not.

#### func  WithIsStatusCodeAnError

```go
func WithIsStatusCodeAnError(f IsStatusCodeAnErrorFunc) Option
```
WithIsStatusCodeAnError customizes the function used for deciding whether a
given call was an error
