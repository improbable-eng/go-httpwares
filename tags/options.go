// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package httpwares_ctxtags

import "net/http"

var (
	defaultOptions = &options{
		tagExtractors: []RequestTagExtractorFunc{},
	}
)

type options struct {
	tagExtractors []RequestTagExtractorFunc
}

func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

type Option func(*options)


// RequestTagExtractorFunc is a signature of user-customizeable functions for extracting tags from requests.
type RequestTagExtractorFunc func (req *http.Request) map[string]interface{}

// WithTagExtractor adds another request tag extractor, allowing you to customize what tags get prepopulated from the request.
func WithTagExtractor(f RequestTagExtractorFunc) Option {
	return func(o *options) {
		o.tagExtractors = append(o.tagExtractors, f)
	}
}
