package minlog_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	pkgErrors "github.com/pkg/errors"

	"github.com/michurin/minlog"
)

func constTime() time.Time {
	return time.Unix(186777777, 777000000).UTC()
}

var withConstTime = minlog.WithNower(constTime)

func ExampleLog_levelInfoAndLabels() {
	l := minlog.New(withConstTime)
	ctx := context.Background()
	ctx = minlog.Label(ctx, "component-a")
	ctx = minlog.Label(ctx, "request-75")
	l.Log(ctx, "just string", true, []byte("valid utf8"), []byte{0xff})
	// Output:
	// 1975-12-02 18:42:57 info component-a:request-75 example_test.go:27 just string true valid utf8 [255]
}

func ExampleLog_levelError() {
	l := minlog.New(withConstTime)
	ctx := context.Background()
	l.Log(ctx, "Error:", errors.New("diagnostics"))
	// Output:
	// 1975-12-02 18:42:57 error example_test.go:35 Error: diagnostics
}

func ExampleLog_nilContext() {
	l := minlog.New(withConstTime)
	l.Log(nil, "Error:", errors.New("diagnostics"))
	// Output:
	// 1975-12-02 18:42:57 error example_test.go:42 Error: diagnostics
}

func Example_pkgErrorsCompatMultilineJSONEncoding() {
	l := minlog.New(
		withConstTime,
		minlog.WithLineFormatter(func(tm, level, label, caller, msg string) string {
			b, _ := json.Marshal(map[string]string{"time": tm, "level": level, "label": label, "caller": caller, "msg": msg})
			return string(b)
		}))
	ctx := context.Background()
	err := errors.New("diagnostics")
	err = pkgErrors.WithMessage(err, "additional message")
	err = pkgErrors.WithMessage(err, "more details")
	l.Log(ctx, "Error:", err)
	// Output:
	// {"caller":"example_test.go:58","label":"","level":"error","msg":"Error: diagnostics\nadditional message\nmore details","time":"1975-12-02 18:42:57"}
}

func ExampleLabel_context() {
	l := minlog.New(withConstTime)
	ctx := minlog.Label(context.Background(), "scope")
	l.Log(ctx, "ok")
	// Output:
	// 1975-12-02 18:42:57 info scope example_test.go:66 ok
}

func ExampleLabel_nestedContext() {
	l := minlog.New(withConstTime)
	ctx := context.Background()
	ctx = minlog.Label(ctx, "scope")
	ctx = minlog.Label(ctx, "subscope")
	l.Log(ctx, "ok")
	// Output:
	// 1975-12-02 18:42:57 info scope:subscope example_test.go:76 ok
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
	// 1975-12-02 18:42:57 [info] example_test.go:89 [component-a] "ok"
}

func ExampleWithTimeFormat() {
	l := minlog.New(
		minlog.WithNower(constTime),
		minlog.WithTimeFormat(time.RFC3339Nano),
	)
	ctx := context.Background()
	l.Log(ctx, "ok")
	// Output:
	// 1975-12-02T18:42:57.777Z info example_test.go:100 ok
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
	// "1975-12-02 18:42:57 info example_test.go:112 ok\n"
}

func ExampleWithLabelPlaceholder() {
	l := minlog.New(
		minlog.WithNower(constTime),
		minlog.WithLabelPlaceholder("<nolabel>"),
	)
	ctx := context.Background()
	l.Log(ctx, "ok")
	// Output:
	// 1975-12-02 18:42:57 info <nolabel> example_test.go:124 ok
}

func ExampleWithLevelLabels() {
	l := minlog.New(
		minlog.WithNower(constTime),
		minlog.WithLevelLabels("[INFO_]", "[ERROR]"),
	)
	ctx := context.Background()
	l.Log(ctx, "ok")
	l.Log(ctx, errors.New("error details"))
	// Output:
	// 1975-12-02 18:42:57 [INFO_] example_test.go:135 ok
	// 1975-12-02 18:42:57 [ERROR] example_test.go:136 error details
}

func ExampleSetDefaultLogger() {
	minlog.SetDefaultLogger(minlog.New(withConstTime))
	minlog.Log(context.Background(), "ok")
	// Output:
	// 1975-12-02 18:42:57 info example_test.go:144 ok
}
