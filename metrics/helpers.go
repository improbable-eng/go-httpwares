// Copyright 2017 Mark Nevill. All Rights Reserved.
// See LICENSE for licensing terms.

package http_metrics

import (
	"io"
	"net/http"
)

type body struct {
	parent io.ReadCloser
	size   int
	done   func(int)
}

func (b *body) Read(p []byte) (int, error) {
	n, err := b.parent.Read(p)
	b.size += n
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		b.done(b.size)
		b.done = func(int) {}
	}
	return n, err
}

func (b *body) Close() error {
	b.done(b.size)
	return b.parent.Close()
}

type writer struct {
	parent  http.ResponseWriter
	started func(int)
	status  int
	size    int
}

func (w *writer) Header() http.Header {
	return w.parent.Header()
}

func (w *writer) WriteHeader(status int) {
	if w.started != nil {
		w.status = status
		w.started(status)
		w.started = nil
	}
	w.parent.WriteHeader(status)
}

func (w *writer) Write(buf []byte) (int, error) {
	if w.started != nil {
		w.status = http.StatusOK
		w.started(w.status)
		w.started = nil
	}
	n, err := w.parent.Write(buf)
	w.size += n
	return n, err
}
