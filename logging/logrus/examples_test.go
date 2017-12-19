package http_logrus

import (
	"github.com/improbable-eng/go-httpwares"
	"github.com/improbable-eng/go-httpwares/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
	"net/http"
)

func ExampleMiddleware() {
	h := http.NewServeMux()
	h.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctxlogrus.Extract(r.Context()).Info("logging")
	})
	hm := Middleware(logrus.WithField("test", "abc"))(h)
	if err := http.ListenAndServe(":8080", hm); err != nil {
		panic(err)
	}
}

func ExampleWithRequestFieldExtractor() {
	Middleware(logrus.WithField("foo", "bar"),
		WithRequestFieldExtractor(func(req *http.Request) map[string]interface{} {
			return map[string]interface{}{
				"http.request.customFieldA": "test",
				"http.request.customFieldB": true,
			}
		}),
	)
}

func ExampleWithResponseFieldExtractor() {
	Middleware(logrus.WithField("foo", "bar"),
		WithResponseFieldExtractor(func(res httpwares.WrappedResponseWriter, req *http.Request) map[string]interface{} {
			return map[string]interface{}{
				"http.response.customFieldC": 1234,
				"http.response.customFieldD": "blah",
			}
		}),
	)
}
