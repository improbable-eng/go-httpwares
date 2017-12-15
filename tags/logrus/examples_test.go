// Copyright (c) Improbable Worlds Ltd, All Rights Reserved

package ctx_logrus_test

import (
	"context"

	"github.com/improbable-eng/go-httpwares/tags"
	"github.com/improbable-eng/go-httpwares/tags/logrus"
)

// Simple unary handler that adds custom fields to the requests's context. These will be used for all log statements.
func ExampleExtract() {
	ctx := context.Background()
	// Add fields the ctxtags of the request which will be added to all extracted loggers.
	http_ctxtags.Extract(ctx).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)
	// Extract a single request-scoped logrus.Logger and log messages.
	l := ctx_logrus.Extract(ctx)
	l.Info("some log message")
	l.Info("another log message")
}
