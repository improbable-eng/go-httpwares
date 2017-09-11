// Copyright 2017 Mark Nevill. All Rights Reserved.
// See LICENSE for licensing terms.

package http_metrics

import (
	"io"
	"net/http"
	"testing"
)

func TestWrappers_ImplementExpectedInterfaces(t *testing.T) {
	var _ io.WriterTo = bodyWT{}
	var _ http.Flusher = writerF{}
	var _ http.Flusher = writerFRF{}
	var _ io.ReaderFrom = writerRF{}
	var _ io.ReaderFrom = writerFRF{}
}
