// Copyright 2017 Mark Nevill. All Rights Reserved.
// See LICENSE for licensing terms.

/*
`reporter` provides client and server side reporting of HTTP request and response for stats.

The middleware (server-side) or tripperware (client-side) must be given a reporter to record the stats for each request.

Example implementations:

"github.com/mwitkow/go-httpwares/metrics/prometheus":
  Prometheus-based reporter implementations for client and server metrics. The user may choose what level of
  detail is included using options to these reporters.
*/
package http_reporter
