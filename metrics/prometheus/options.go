// Copyright 2017 Mark Nevill. All Rights Reserved.
// See LICENSE for licensing terms.

package http_prometheus

type options struct {
	name    string
	latency bool
	paths   bool
	hosts   bool
	sizes   bool
}

type opt func(*options)

func evalOpts(opts []opt) *options {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func WithName(name string) opt {
	return func(o *options) {
		o.name = name
	}
}

func WithLatency() opt {
	return func(o *options) {
		o.latency = true
	}
}

func WithHostLabel() opt {
	return func(o *options) {
		o.hosts = true
	}
}

func WithPathLabel() opt {
	return func(o *options) {
		o.paths = true
	}
}

func WithSizes() opt {
	return func(o *options) {
		o.sizes = true
	}
}
