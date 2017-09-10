// Copyright 2017 Mark Nevill. All Rights Reserved.
// See LICENSE for licensing terms.

package http_metrics

import (
	"net/http"
	"time"
)

// Called when a new request is to be tracked.
type Reporter interface {
	// Start tracking a new request.
	Track(req *http.Request) Tracker
}

// Receives events about a tracked request.
type Tracker interface {
	// The exchange has started. This is called immediately after Reporter.Track.
	// On the client, this is called before any data is sent.
	// On the server, this is called after headers have been parsed.
	RequestStarted()
	// The request body has been read to EOF or closed, whichever comes first.
	// On the client, this is called when the transport completes sending the request.
	// On the server, this is called when the handler completes reading the request, and may be omitted.
	RequestRead(duration time.Duration, size int)
	// The handling of the response has started.
	// On the client, this is called after the response headers have been parsed.
	// On the server, this is called before any data is written.
	ResponseStarted(duration time.Duration, status int, header http.Header)
	// The response has completed.
	// On the client, this is called when the body is read to EOF or closed, whichever comes first, and may be omitted.
	// On the server, this is called when the handler returns and has therefore completed writing the response.
	ResponseDone(duration time.Duration, status int, size int)
}
