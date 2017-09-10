// Copyright 2017 Mark Nevill. All Rights Reserved.
// See LICENSE for licensing terms.

/*
`http_metrics` provides client and server side reporting of HTTP stats.

The middleware (server-side) or tripperware (client-side) must be given a reporter to record the stats for each request.

Prometheus-based reporter implementations for client and server metrics are included. The user may choose what level of
detail is included using options to these reporters.
*/
package http_metrics
