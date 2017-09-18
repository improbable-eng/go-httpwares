// Copyright 2017 Mark Nevill. All Rights Reserved.
// See LICENSE for licensing terms.

/*
`reporter` provides a way to listen on particular HTTP request and response events.
It also provides the reporter middleware (server-side) or tripperware (client-side) that sets up all callbacks in place.

Example implementations:
"github.com/mwitkow/go-httpwares/metrics/prometheus":
  Prometheus-based reporter implementations for client and server metrics. The user may choose what level of
  detail is included using options to these reporters.
*/
package http_reporter
