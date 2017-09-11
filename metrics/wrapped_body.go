// Copyright 2017 Mark Nevill. All Rights Reserved.
// See LICENSE for licensing terms.

package http_metrics

import "io"

func wrapBody(rc io.ReadCloser, done func(int)) io.ReadCloser {
	wrapped := &body{
		parent: rc,
		done:   done,
	}
	if wt, ok := rc.(io.WriterTo); ok {
		return bodyWT{body: wrapped, wt: wt}
	}
	return wrapped
}

type body struct {
	parent io.ReadCloser
	size   int
	done   func(int)
}

type bodyWT struct {
	*body
	wt io.WriterTo
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

func (b bodyWT) WriteTo(w io.Writer) (int64, error) {
	n, err := b.wt.WriteTo(w)
	b.body.size += int(n)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		b.done(b.size)
		b.done = func(int) {}
	}
	return n, err
}
