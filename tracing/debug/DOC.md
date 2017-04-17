# http_debug
--
    import "github.com/mwitkow/go-httpwares/tracing/debug"


## Usage

#### func  DefaultStatusCodeIsError

```go
func DefaultStatusCodeIsError(statusCode int) bool
```
DefaultStatusCodeIsError defines a function that says whether a given request is
an error based on a code.

#### func  Middleware

```go
func Middleware(opts ...Option) httpwares.Middleware
```
Middleware returns a http.Handler middleware that writes inbound requests to
/debug/request.

#### func  Tripperware

```go
func Tripperware(opts ...Option) httpwares.Tripperware
```
Tripperware returns a piece of client-side Tripperware that puts requests on a
status page.

#### type FilterFunc

```go
type FilterFunc func(req *http.Request) bool
```

FilterFunc allows users to provide a function that filters out certain methods
from being traced.

If it returns false, the given request will not be traced.

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

#### func  WithStatusCodeIsError

```go
func WithStatusCodeIsError(f StatusCodeIsError) Option
```
WithStatusCodeIsError customizes the function used for deciding whether a given
call was an error

#### type StatusCodeIsError

```go
type StatusCodeIsError func(statusCode int) bool
```

StatusCodeIsError allows the customization of which requests are considered
errors in the tracing system.
