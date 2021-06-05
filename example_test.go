package minlog_test

import (
	"context"
	"errors"
	"time"

	"github.com/michurin/minlog"
)

func ExampleSimplest() {
	l := minlog.New(minlog.WithNower(func() time.Time { return time.Unix(0, 0).UTC() }))
	ctx := context.Background()
	l.Log(ctx, "ok")
	// Output:
	// 1970-01-01 00:00:00 info . example_test.go:14 ok
}

func ExampleSimleContext() {
	l := minlog.New(minlog.WithNower(func() time.Time { return time.Unix(0, 0).UTC() }))
	ctx := minlog.Label(context.Background(), "scope")
	l.Log(ctx, "ok")
	// Output:
	// 1970-01-01 00:00:00 info scope example_test.go:22 ok
}

func ExampleNestedContext() {
	l := minlog.New(minlog.WithNower(func() time.Time { return time.Unix(0, 0).UTC() }))
	ctx := context.Background()
	ctx = minlog.Label(ctx, "scope")
	ctx = minlog.Label(ctx, "subscope")
	l.Log(ctx, "ok")
	// Output:
	// 1970-01-01 00:00:00 info scope:subscope example_test.go:32 ok
}

func ExampleError() {
	l := minlog.New(minlog.WithNower(func() time.Time { return time.Unix(0, 0).UTC() }))
	ctx := context.Background()
	l.Log(ctx, "Error:", errors.New("diagnostics"))
	// Output:
	// 1970-01-01 00:00:00 error . example_test.go:40 Error: diagnostics
}

func ExampleCustomLineFormat() {
	l := minlog.New(
		minlog.WithNower(func() time.Time { return time.Unix(0, 0).UTC() }),
		minlog.WithLineFormat("%[1]s [%[2]s] %[4]s [%[3]s] %[5]q"),
	)
	ctx := minlog.Label(context.Background(), "component-a")
	l.Log(ctx, "ok")
	// Output:
	// 1970-01-01 00:00:00 [info] example_test.go:51 [component-a] "ok"
}

func ExampleSetAndUseGlobalLogger() {
	minlog.SetDefaultLogger(minlog.New(minlog.WithNower(func() time.Time { return time.Unix(0, 0).UTC() })))
	minlog.Log(context.Background(), "ok")
	// Output:
	// 1970-01-01 00:00:00 info . example_test.go:58 ok
}
