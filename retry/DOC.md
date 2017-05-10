# http_retry
--
    import "github.com/mwitkow/go-httpwares/retry"

`http_retry` is a HTTP client-side Tripperware that allows you to retry requests
that are marked as idempotent and safe.

This logic works for requests considered safe and idempotent (configurable) and
ones that have no body, or have `Request.GetBody` either automatically
implemented (`byte.Buffer`, `string.Buffer`) or specified manually. By default
all GET, HEAD and OPTIONS requests are considered idempotent and safe.

### The implementation allow retries after client-side errors or returned error
codes (by default 5**) according to configurable backoff policies (linear
backoff by default). Additionally, requests can be hedged. Hedging works by
sending additional requests without waiting for a previous one to return.

## Usage

#### func  DefaultResponseDiscarder

```go
func DefaultResponseDiscarder(resp *http.Response) bool
```
DefaultResponseDiscarder is the default implementation that discards responses
in order to try again.

It is fairly conservative and rejects (and thus retries) responses with 500, 503
and 504 status codes. See
https://en.wikipedia.org/wiki/List_of_HTTP_status_codes#5xx_Server_error

#### func  DefaultRetriableDecider

```go
func DefaultRetriableDecider(req *http.Request) bool
```
DefaultRetriableDecider is the default implementation that retries only
indempotent and safe requests (GET, OPTION, HEAD).

It is fairly conservative and heeds the of
http://restcookbook.com/HTTP%20Methods/idempotency.

#### func  Enable

```go
func Enable(req *http.Request) *http.Request
```
Enable turns on the retry logic for a given request, regardless of what the
retry decider says.

Please make sure you do not pass around this request's context.

#### func  EnableContext

```go
func EnableContext(ctx context.Context) context.Context
```
Enable turns on the retry logic for a given request's context, regardless of
what the retry decider says.

Please make sure you do not pass around this request's context.

#### func  Tripperware

```go
func Tripperware(opts ...Option) httpwares.Tripperware
```
Tripperware is client side HTTP ware that retries the requests.

Be default this retries safe and idempotent requests 3 times with a linear delay
of 100ms. This behaviour can be customized using With* parameter options.

Requests that have `http_retry.Enable` set on them will always be retried.

#### type BackoffFunc

```go
type BackoffFunc func(attempt uint) time.Duration
```

BackoffFunc denotes a family of functions that controll the backoff duration
between call retries.

They are called with an identifier of the attempt, and should return a time the
system client should hold off for. If the time returned is longer than the
`context.Context.Deadline` of the request the deadline of the request takes
precedence and the wait will be interrupted before proceeding with the next
iteration.

#### func  BackoffLinear

```go
func BackoffLinear(waitBetween time.Duration) BackoffFunc
```
BackoffLinear is very simple: it waits for a fixed period of time between calls.

#### type Option

```go
type Option func(*options)
```


#### func  WithBackoff

```go
func WithBackoff(bf BackoffFunc) Option
```
WithBackoff sets the `BackoffFunc `used to control time between retries.

#### func  WithDecider

```go
func WithDecider(f RequestRetryDeciderFunc) Option
```
WithDecider is a function that allows users to customize the logic that decides
whether a request is retriable.

#### func  WithMax

```go
func WithMax(maxRetries uint) Option
```
WithMax sets the maximum number of retries on this call, or this interceptor.

#### func  WithResponseDiscarder

```go
func WithResponseDiscarder(f RequestRetryDeciderFunc) Option
```
WithResponseDiscarder is a function that decides whether a given response should
be discarded and another request attempted.

#### type RequestRetryDeciderFunc

```go
type RequestRetryDeciderFunc func(req *http.Request) bool
```

RequestRetryDeciderFunc decides whether the given function is idempotent and
safe or to retry.

#### type ResponseDiscarderFunc

```go
type ResponseDiscarderFunc func(resp *http.Response) bool
```

ResponseDiscarderFunc decides when to discard a response and retry the request
again (on true).
