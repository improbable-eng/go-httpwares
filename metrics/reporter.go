// Copyright 2017 Mark Nevill. All Rights Reserved.
// See LICENSE for licensing terms.

package http_metrics

import (
	"net/http"
	"time"
)

type Reporter interface {
	Track(req *http.Request) Tracker
}

type Tracker interface {
	RequestStarted()
	RequestRead(duration time.Duration, size int)
	ResponseStarted(duration time.Duration, status int, header http.Header)
	ResponseDone(duration time.Duration, status int, size int)
}
