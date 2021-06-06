package minlog_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/michurin/minlog"
)

func constTime() time.Time {
	return time.Unix(186777777, 777000000).UTC()
}

var withConstTime = minlog.WithNower(constTime)

func Example_levelInfo() {
	l := minlog.New(withConstTime)
	ctx := context.Background()
	ctx = minlog.Label(ctx, "component-a")
	ctx = minlog.Label(ctx, "request-75")
	l.Log(ctx, "ok")
	// Output:
	// 1975-12-02 18:42:57 info component-a:request-75 example_test.go:24 ok
}

func Example_levelError() {
	l := minlog.New(withConstTime)
	ctx := context.Background()
	l.Log(ctx, "Error:", errors.New("diagnostics"))
	// Output:
	// 1975-12-02 18:42:57 error example_test.go:32 Error: diagnostics
}

func ExampleLabel_context() {
	l := minlog.New(withConstTime)
	ctx := minlog.Label(context.Background(), "scope")
	l.Log(ctx, "ok")
	// Output:
	// 1975-12-02 18:42:57 info scope example_test.go:40 ok
}

func ExampleLabel_nestedContext() {
	l := minlog.New(withConstTime)
	ctx := context.Background()
	ctx = minlog.Label(ctx, "scope")
	ctx = minlog.Label(ctx, "subscope")
	l.Log(ctx, "ok")
	// Output:
	// 1975-12-02 18:42:57 info scope:subscope example_test.go:50 ok
}

func ExampleWithLineFormatter() {
	l := minlog.New(
		minlog.WithNower(constTime),
		minlog.WithLineFormatter(func(tm, level, label, caller, msg string) string {
			return fmt.Sprintf("%[1]s [%[2]s] %[4]s [%[3]s] %[5]q", tm, level, label, caller, msg)
		}),
	)
	ctx := minlog.Label(context.Background(), "component-a")
	l.Log(ctx, "ok")
	// Output:
	// 1975-12-02 18:42:57 [info] example_test.go:63 [component-a] "ok"
}

func ExampleWithTimeFormat() {
	l := minlog.New(
		minlog.WithNower(constTime),
		minlog.WithTimeFormat(time.RFC3339Nano),
	)
	ctx := minlog.Label(context.Background(), "component-a")
	l.Log(ctx, "ok")
	// Output:
	// 1975-12-02T18:42:57.777Z info component-a example_test.go:74 ok
}

func ExampleWithWriter() {
	output := new(bytes.Buffer)
	l := minlog.New(
		minlog.WithNower(constTime),
		minlog.WithWriter(output),
	)
	ctx := context.Background()
	l.Log(ctx, "ok")
	fmt.Printf("%q\n", output.String())
	// Output:
	// "1975-12-02 18:42:57 info example_test.go:86 ok\n"
}

func ExampleSetDefaultLogger() {
	minlog.SetDefaultLogger(minlog.New(withConstTime))
	minlog.Log(context.Background(), "ok")
	// Output:
	// 1975-12-02 18:42:57 info example_test.go:94 ok
}
