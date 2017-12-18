package http_logrus_test

import (
	"net/http"

	"github.com/improbable-eng/go-httpwares/tags"
	"github.com/improbable-eng/go-httpwares/tags/logrus"
)

var handler http.HandlerFunc

// Simple example of a `http.Handler` extracting the `Middleware`-injected logrus logger from the context.
func ExampleExtract_withCustomTags() {
	handler = func(resp http.ResponseWriter, req *http.Request) {
		// Handlers can add extra tags to `http_ctxtags` that will be set in both the extracted loggers *and*
		// the final log statement.
		http_ctxtags.ExtractInbound(req).Set("my_custom.my_string", "something").Set("my_custom.my_int", 1337)
		ctx_logrus.Extract(req).Warningf("Hello World")
	}
}
