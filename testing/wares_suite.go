// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package httpwares_testing

import (
	"net"
	"time"

	"flag"
	"path"
	"runtime"

	"crypto/tls"
	"net/http"

	"github.com/mwitkow/go-conntrack/connhelpers"
	"github.com/mwitkow/go-httpwares"
	"github.com/pressly/chi"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"
)

var (
	flagTls = flag.Bool("use_tls", true, "whether all gRPC middleware tests should use tls")
)

func getTestingCertsPath() string {
	_, callerPath, _, _ := runtime.Caller(0)
	return path.Join(path.Dir(callerPath), "certs")
}

// WaresTestSuite is a testify/Suite that starts a gRPC PingService server and a client.
type WaresTestSuite struct {
	suite.Suite

	ServerMiddleware  []httpwares.Middleware
	ClientTripperware httpwares.TripperwareChain

	Handler http.Handler

	ServerListener net.Listener
	Server         *http.Server
}

func (s *WaresTestSuite) SetupSuite() {
	var err error
	if s.ServerListener == nil {
		s.ServerListener, err = net.Listen("tcp", "127.0.0.1:0")
		require.NoError(s.T(), err, "must be able to allocate a port for serverListener")
		if *flagTls {
			tlsConf, err := connhelpers.TlsConfigForServerCerts(
				path.Join(getTestingCertsPath(), "localhost.crt"),
				path.Join(getTestingCertsPath(), "localhost.key"),
			)
			require.NoError(s.T(), err, "failed starting TLS config for WaresTestSuite")
			tlsConf, err = connhelpers.TlsConfigWithHttp2Enabled(tlsConf)
			s.ServerListener = tls.NewListener(s.ServerListener, tlsConf)
		}
	}
	if s.Handler == nil {
		s.Handler = http.HandlerFunc(PingBackHandler(DefaultPingBackStatusCode))
	}
	if s.Server == nil {
		handler := s.Handler
		if len(s.ServerMiddleware) > 0 {
			chains := [](func(http.Handler) http.Handler){}
			for _, ware := range s.ServerMiddleware {
				chains = append(chains, ware)
			}
			handler = chi.Chain(chains...).Handler(handler)
		}
		s.Server = &http.Server{
			ErrorLog: nil, // TODO(mwitkow): Add ErrorLog to testint.T.Log
			Handler:  handler,
		}
	}

	go func() {
		s.Server.Serve(s.ServerListener)
	}()
}

// NewClient returns a client that dials the server for *any* address provided. It's up to you to validate that the
// scheme is right.
func (s *WaresTestSuite) NewClient() *http.Client {
	var transport http.RoundTripper = &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.Dial(network, s.ServerAddr())
		},
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	if s.ClientTripperware != nil {
		transport = s.ClientTripperware.Forge(transport)
	}
	return &http.Client{
		Transport: transport,
	}
}

func (s *WaresTestSuite) ServerAddr() string {
	return s.ServerListener.Addr().String()
}

func (s *WaresTestSuite) SimpleCtx() context.Context {
	ctx, _ := context.WithTimeout(context.TODO(), 1*time.Second)
	return ctx
}

func (s *WaresTestSuite) TearDownSuite() {
	time.Sleep(10 * time.Millisecond)
	if s.ServerListener != nil {
		s.Server.Close()
		s.ServerListener.Close()
	}
}
