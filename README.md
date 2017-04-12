# :stew: Go HTTP Wares 

[![Travis Build](https://travis-ci.org/mwitkow/go-httpwares.svg?branch=master)](https://travis-ci.org/mwitkow/go-httpwares)
[![Go Report Card](https://goreportcard.com/badge/github.com/mwitkow/go-httpwares)](https://goreportcard.com/report/github.com/mwitkow/go-httpwares)
[![GoDoc](http://img.shields.io/badge/GoDoc-Reference-blue.svg)](https://godoc.org/github.com/mwitkow/go-httpwares)
[![SourceGraph](https://sourcegraph.com/github.com/mwitkow/go-httpwares/-/badge.svg)](https://sourcegraph.com/github.com/mwitkow/go-httpwares/?badge)
[![codecov](https://codecov.io/gh/mwitkow/go-httpwares/branch/master/graph/badge.svg)](https://codecov.io/gh/mwitkow/go-httpwares)
[![Apache 2.0 License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![quality: alpha](https://img.shields.io/badge/quality-alpha-orange.svg)](#status)

Context-based Client and Server middleware libraries.

The libraries are meant to rely only on the standard `http` package functions and make heavy use of `context` (introduced in Go 1.7) for propagating state between handlers. The libraries come in two flavours:
 * server-side `Middleware` (`func (http.Handler) http.Handler)`) 
 * client-side `Tripperware` (`func (http.RoundTripper) http.RoundTripper`) 

These are meant as excellent companions for interceptors of [`github.com/grpc-ecosystem/go-grpc-middleware`](https://github.com/grpc-ecosystem/go-grpc-middleware) making it easy to build combined gRPC/HTTP Golang servers.

## Why?

Having a consistent set of inbound and outbound handlers make it easy to log, trace, auth and debug your handlers.
