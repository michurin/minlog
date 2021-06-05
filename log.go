package minlog

import (
	"context"
	"fmt"
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
	fileNameCutter func(string) string
	formatter      func(...interface{}) (string, string)
	lineFmt        string
	callerLevel    int
	// TODO add output stream
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
		formatter:      simpleFormatter,
		lineFmt:        "%s %s %s %s %s",
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
	if label == "" {
		label = "."
	}
	level, msg := l.formatter(message...)
	fmt.Printf(l.lineFmt, tm, level, label, caller, msg)
}

func (l *Logger) caller() string {
	_, file, line, ok := runtime.Caller(l.callerLevel)
	if !ok {
		return "[no file]"
	}
	return fmt.Sprintf("%s:%d", l.fileNameCutter(file), line)
}

// ------------

var defaultLogger Interface

func Log(ctx context.Context, message ...interface{}) {
	defaultLogger.Log(ctx, message...)
}

func SetDefaultLogger(l *Logger) {
	l.callerLevel++
	defaultLogger = l
}

// ------------

type Option func(*Logger)

func WithTimeFormat(fmt string) Option {
	return func(l *Logger) {
		l.timeFmt = fmt
	}
}

func WithLineFormat(fmt string) Option {
	return func(l *Logger) {
		l.lineFmt = fmt
	}
}

func WithNower(nwr func() time.Time) Option {
	return func(l *Logger) {
		l.nower = nwr
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

//

func simpleFormatter(mm ...interface{}) (string, string) {
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

// -- context

func Label(ctx context.Context, label string) context.Context {
	l, _ := ctx.Value(labelKey).(string)
	if l != "" {
		label = l + ":" + label
	}
	return context.WithValue(ctx, labelKey, label)
}
