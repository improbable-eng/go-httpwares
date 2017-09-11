// Copyright 2017 Mark Nevill. All Rights Reserved.
// See LICENSE for licensing terms.

package http_metrics

import (
	"io"
	"net/http"
)

type wrappedWriter interface {
	http.ResponseWriter
	Status() int
	Size() int
}

func wrapWriter(w http.ResponseWriter, started func(int)) wrappedWriter {
	wrapped := &writer{
		parent:  w,
		started: started,
	}
	f, isFlusher := w.(http.Flusher)
	rf, isReaderFrom := w.(io.ReaderFrom)
	if isFlusher && isReaderFrom {
		return writerFRF{writer: wrapped, f: f, rf: rf}
	}
	if isFlusher {
		return writerF{writer: wrapped, f: f}
	}
	if isReaderFrom {
		return writerRF{writer: wrapped, rf: rf}
	}
	return wrapped
}

type writer struct {
	parent  http.ResponseWriter
	started func(int)
	status  int
	size    int
}

type writerF struct {
	*writer
	f http.Flusher
}

type writerRF struct {
	*writer
	rf io.ReaderFrom
}

type writerFRF struct {
	*writer
	f  http.Flusher
	rf io.ReaderFrom
}

func (w *writer) Status() int {
	return w.status
}

func (w *writer) Size() int {
	return w.size
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

func (w writerF) Flush() {
	w.f.Flush()
}

func (w writerFRF) Flush() {
	w.f.Flush()
}

func (w writerRF) ReadFrom(r io.Reader) (int64, error) {
	n, err := w.rf.ReadFrom(r)
	w.size += int(n)
	return n, err
}

func (w writerFRF) ReadFrom(r io.Reader) (int64, error) {
	n, err := w.rf.ReadFrom(r)
	w.size += int(n)
	return n, err
}
