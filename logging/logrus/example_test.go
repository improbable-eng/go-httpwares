package http_logrus

import (
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
