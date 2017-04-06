# httpwares_ctxtags
--
    import "github.com/mwitkow/go-httpwares/tags"


## Usage

#### func  ChiRouteTagExtractor

```go
func ChiRouteTagExtractor(req *http.Request) map[string]interface{}
```
ChiRouteTagExtractor extracts chi router information and puts them into tags.

#### func  Middleware

```go
func Middleware(opts ...Option) httpwares.Middleware
```
Middleware returns a http.Handler middleware values for request tags.

#### type Option

```go
type Option func(*options)
```


#### func  WithTagExtractor

```go
func WithTagExtractor(f RequestTagExtractorFunc) Option
```
WithTagExtractor adds another request tag extractor, allowing you to customize
what tags get prepopulated from the request.

#### type RequestTagExtractorFunc

```go
type RequestTagExtractorFunc func(req *http.Request) map[string]interface{}
```

RequestTagExtractorFunc is a signature of user-customizeable functions for
extracting tags from requests.

#### type Tags

```go
type Tags struct {
}
```

Tags is the struct used for storing request tags between Context calls. This
object is *not* thread safe, and should be handled only in the context of the
request.

#### func  Extract

```go
func Extract(req *http.Request) *Tags
```
Extracts returns a pre-existing Tags object in the request's Context. If the
context wasn't set in a tag interceptor, a no-op Tag storage is returned that
will *not* be propagated in context.

#### func  ExtractFromContext

```go
func ExtractFromContext(ctx context.Context) *Tags
```
Extracts returns a pre-existing Tags object in the request's Context. If the
context wasn't set in a tag interceptor, a no-op Tag storage is returned that
will *not* be propagated in context.

#### func (*Tags) Has

```go
func (t *Tags) Has(key string) bool
```
Has checks if the given key exists.

#### func (*Tags) Set

```go
func (t *Tags) Set(key string, value interface{}) *Tags
```
Set sets the given key in the metadata tags.

#### func (*Tags) Values

```go
func (t *Tags) Values() map[string]interface{}
```
Values returns a map of key to values. Do not modify the underlying map, please
use Set instead.
