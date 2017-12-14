package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/improbable-eng/go-httpwares"
	"github.com/improbable-eng/go-httpwares/logging/logrus"
	"github.com/improbable-eng/go-httpwares/tags"
	"github.com/improbable-eng/go-httpwares/tracing/debug"
	"github.com/improbable-eng/go-httpwares/tracing/opentracing"
	"github.com/mwitkow/go-conntrack"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context/ctxhttp"
	_ "golang.org/x/net/trace" // import the debug pages
)

var (
	port = flag.Int("port", 9090, "whether to use tls or not")
)

func main() {
	flag.Parse()

	log.SetOutput(os.Stdout)
	logInstance := log.NewEntry(log.StandardLogger())
	opentracing.SetGlobalTracer(mocktracer.New())

	googleClient := httpwares.WrapClient(
		http.DefaultClient,
		http_ctxtags.Tripperware(),
		http_opentracing.Tripperware(),
		http_debug.Tripperware(),
	)

	handlerFunc := func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(http.StatusOK)
		resp.Header().Add("Content-Type", "application/json")
		resp.Write([]byte(`{"queried_google": true}`))
		callCtx := conntrack.DialNameToContext(req.Context(), "google")
		_, err := ctxhttp.Get(callCtx, googleClient, "https://www.google.com")
		logInstance.Printf("Google reached with err: %v", err)
	}

	chainedHandler := chi.Chain(
		http_ctxtags.Middleware("google_ping_service"),
		http_opentracing.Middleware(),
		http_debug.Middleware(),
		http_logrus.Middleware(logInstance),
	).HandlerFunc(handlerFunc)

	http.DefaultServeMux.Handle("/", chainedHandler)
	http.DefaultServeMux.Handle("/metrics", prometheus.Handler())

	httpServer := http.Server{
		Addr:     fmt.Sprintf(":%d", *port),
		Handler:  http.DefaultServeMux,
		ErrorLog: http_logrus.AsHttpLogger(logInstance),
	}

	log.Printf("Listening on http://localhost:%d/", *port)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("Failed listning: %v", err)
	}
}
