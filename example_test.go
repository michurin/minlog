package minlog_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"time"

	pkgErrors "github.com/pkg/errors"

	"github.com/michurin/minlog"
)

var (
	constTime            = time.Unix(186777777, 777000000).UTC()
	withReproducibleTime = minlog.WithNower(func() time.Time { return constTime })
)

func ExampleLog_levelInfoAndLabels() {
	l := minlog.New(withReproducibleTime)
	ctx := context.Background()
	ctx = minlog.Label(ctx, "component-a")
	ctx = minlog.Label(ctx, "request-75")
	l.Log(ctx, "just string", true, []byte("valid utf8"), []byte{0xff})
	// Output:
	// 1975-12-02 18:42:57 info component-a:request-75 example_test.go:27 just string true valid utf8 [255]
}

func ExampleLog_levelError() {
	l := minlog.New(withReproducibleTime)
	ctx := context.Background()
	l.Log(ctx, "Error:", errors.New("diagnostics"))
	// Output:
	// 1975-12-02 18:42:57 error example_test.go:35 Error: diagnostics
}

func ExampleLog_nilContext() {
	l := minlog.New(withReproducibleTime)
	l.Log(nil, "Error:", errors.New("diagnostics")) //nolint:staticcheck // disable nil context warning
	// Output:
	// 1975-12-02 18:42:57 error example_test.go:42 Error: diagnostics
}

func Example_pkgErrorsCompatMultilineJSONEncoding() {
	l := minlog.New(
		withReproducibleTime,
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
	l := minlog.New(withReproducibleTime)
	ctx := minlog.Label(context.Background(), "scope")
	l.Log(ctx, "ok")
	// Output:
	// 1975-12-02 18:42:57 info scope example_test.go:66 ok
}

func ExampleLabel_nestedContext() {
	l := minlog.New(withReproducibleTime)
	ctx := context.Background()
	ctx = minlog.Label(ctx, "scope")
	ctx = minlog.Label(ctx, "subscope")
	l.Log(ctx, "ok")
	// Output:
	// 1975-12-02 18:42:57 info scope:subscope example_test.go:76 ok
}

func ExampleWithLineFormatter() {
	l := minlog.New(
		withReproducibleTime,
		minlog.WithLineFormatter(func(tm, level, label, caller, msg string) string {
			return fmt.Sprintf("%[1]s [%[2]s] %[4]s [%[3]s] %[5]q", tm, level, label, caller, msg)
		}),
	)
	ctx := minlog.Label(context.Background(), "component-a")
	l.Log(ctx, "ok")
	// Output:
	// 1975-12-02 18:42:57 [info] example_test.go:89 [component-a] "ok"
}

func ExampleWithLabelFormatter() {
	l := minlog.New(
		withReproducibleTime,
		minlog.WithLabelFormatter(func(v interface{}) string {
			return fmt.Sprintf("[%s]", v)
		}),
	)
	ctx := minlog.Label(context.Background(), "component-a")
	l.Log(ctx, "ok")
	// Output:
	// 1975-12-02 18:42:57 info [component-a] example_test.go:102 ok
}

func ExampleWithTimeFormat() {
	l := minlog.New(withReproducibleTime, minlog.WithTimeFormat(time.RFC3339Nano))
	ctx := context.Background()
	l.Log(ctx, "ok")
	// Output:
	// 1975-12-02T18:42:57.777Z info example_test.go:110 ok
}

func ExampleWithWriter() {
	output := new(bytes.Buffer)
	l := minlog.New(withReproducibleTime, minlog.WithWriter(output))
	ctx := context.Background()
	l.Log(ctx, "ok")
	fmt.Printf("%q\n", output.String())
	// Output:
	// "1975-12-02 18:42:57 info example_test.go:119 ok\n"
}

func ExampleWithLabelPlaceholder() {
	l := minlog.New(withReproducibleTime, minlog.WithLabelPlaceholder("<nolabel>"))
	ctx := context.Background()
	l.Log(ctx, "ok")
	// Output:
	// 1975-12-02 18:42:57 info <nolabel> example_test.go:128 ok
}

func ExampleWithCommonLabel() {
	l := minlog.New(withReproducibleTime, minlog.WithCommonLabel("logger_label"))
	ctx := context.Background()
	l.Log(ctx, "ok")
	l.Log(minlog.Label(ctx, "from_ctx"), "ok")
	// Output:
	// 1975-12-02 18:42:57 info logger_label example_test.go:136 ok
	// 1975-12-02 18:42:57 info logger_label:from_ctx example_test.go:137 ok
}

func ExampleWithLevelLabels() {
	l := minlog.New(withReproducibleTime, minlog.WithLevelLabels("[INFO_]", "[ERROR]"))
	ctx := context.Background()
	l.Log(ctx, "ok")
	l.Log(ctx, errors.New("error details"))
	// Output:
	// 1975-12-02 18:42:57 [INFO_] example_test.go:146 ok
	// 1975-12-02 18:42:57 [ERROR] example_test.go:147 error details
}

func ExampleWithCallerCutter() {
	l := minlog.New(withReproducibleTime, minlog.WithCallerCutter(path.Base))
	l.Log(context.Background(), "ok")
	// Output:
	// 1975-12-02 18:42:57 info example_test.go:155 ok
}

func ExampleSetDefaultLogger() {
	minlog.SetDefaultLogger(minlog.New(withReproducibleTime))
	minlog.Log(context.Background(), "ok")
	// Output:
	// 1975-12-02 18:42:57 info example_test.go:162 ok
}
