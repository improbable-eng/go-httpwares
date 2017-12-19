package ctxlogrus_test

import (
	"context"
	"github.com/improbable-eng/go-httpwares/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
)

func ExampleExtract() {
	ctx := context.Background()
	entry := ctxlogrus.Extract(ctx)
	entry.Info("logging")
}

func ExampleToContext() {
	ctx := context.Background()
	ctx = ctxlogrus.ToContext(ctx, logrus.WithFields(logrus.Fields{"foo": "bar"}))
}

func ExampleAddFields() {
	ctx := context.Background()
	ctx = ctxlogrus.ToContext(ctx, logrus.WithFields(logrus.Fields{"foo": "bar"}))
	ctxlogrus.AddFields(ctx, logrus.Fields{"num": 42})
}
