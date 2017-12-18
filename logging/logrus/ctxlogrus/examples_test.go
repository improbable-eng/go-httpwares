package ctxlogrus_test

import (
	"context"
	"github.com/improbable-eng/go-httpwares/logging/logrus/ctxlogrus"
)

func ExampleExtract() {
	ctx := context.Background()
	entry := ctxlogrus.Extract(ctx)
	entry.Info("logging")
}
