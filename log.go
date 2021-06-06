package minlog

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

type Interface interface {
	Log(ctx context.Context, message ...interface{})
}

type Logger struct {
	timeFmt        string
	nower          func() time.Time
	fileNameCutter func(string) string                   // it has to be option WithFileNameCutter?
	formatter      func(...interface{}) (string, string) // it has to be option too?
	lineFormatter  func(tm, level, label, caller, msg string) string
	output         io.Writer
	callerLevel    int
}

type logContextKey string

const labelKey = logContextKey("label")

func New(opt ...Option) *Logger {
	_, file, _, _ := runtime.Caller(1)
	dirname := file[:len(file)-len(path.Base(file))] // including final separator
	l := &Logger{
		timeFmt:        "2006-01-02 15:04:05",
		nower:          time.Now,
		fileNameCutter: mkLongestPrefixCutter(dirname),
		formatter:      defaultFormatter,
		lineFormatter:  defaultLineFomatter,
		output:         os.Stdout,
		callerLevel:    2,
	}
	for _, o := range opt {
		o(l)
	}
	return l
}

func (l *Logger) Log(ctx context.Context, message ...interface{}) {
	tm := l.nower().Format(l.timeFmt)
	caller := l.caller()
	label, _ := ctx.Value(labelKey).(string)
	level, msg := l.formatter(message...)
	fmt.Fprintln(l.output, l.lineFormatter(tm, level, label, caller, msg))
}

func (l *Logger) caller() string {
	_, file, line, _ := runtime.Caller(l.callerLevel)
	return fmt.Sprintf("%s:%d", l.fileNameCutter(file), line)
}

// -- func

var defaultLogger Interface

func Log(ctx context.Context, message ...interface{}) {
	defaultLogger.Log(ctx, message...)
}

func SetDefaultLogger(l Interface) {
	if lg, ok := l.(*Logger); ok {
		lg.callerLevel++
	}
	defaultLogger = l
}

// -- options

type Option func(*Logger)

func WithTimeFormat(fmt string) Option {
	return func(l *Logger) {
		l.timeFmt = fmt
	}
}

func WithLineFormatter(fmtr func(tm, level, label, caller, msg string) string) Option {
	return func(l *Logger) {
		l.lineFormatter = fmtr
	}
}

func WithNower(nwr func() time.Time) Option {
	return func(l *Logger) {
		l.nower = nwr
	}
}

func WithWriter(w io.Writer) Option {
	return func(l *Logger) {
		l.output = w
	}
}

// -- helpers

func mkLongestPrefixCutter(t string) func(string) string {
	x := []byte(t)
	l := len(x)
	return func(o string) string {
		for i, b := range []byte(o) {
			if i >= l {
				return o[i:]
			}
			if x[i] != b {
				return o[i:]
			}
		}
		return ""
	}
}

// -- defaults

func defaultFormatter(mm ...interface{}) (string, string) {
	level := "info"
	pp := []string(nil)
	for _, m := range mm {
		switch e := m.(type) {
		case error:
			level = "error"
			pp = append(pp, e.Error()) // TODO reach errors
		default:
			pp = append(pp, fmt.Sprintf("%v", e)) // TODO if e != ""
		}
	}
	return level, strings.Join(pp, " ")
}

func defaultLineFomatter(tm, level, label, caller, msg string) string {
	a := []string{tm, level}
	if label != "" {
		a = append(a, label)
	}
	a = append(a, caller, msg)
	return strings.Join(a, " ")
}

// -- context

func Label(ctx context.Context, label string) context.Context {
	l, _ := ctx.Value(labelKey).(string)
	if l != "" {
		label = l + ":" + label
	}
	return context.WithValue(ctx, labelKey, label)
}
