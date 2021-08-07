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
	"unicode/utf8"
)

const (
	DefaultInfoLabel  = "info"
	DefaultErrorLabel = "error"

	baseCallerLevel = 2
)

type Interface interface {
	Log(ctx context.Context, message ...interface{})
}

type Logger struct {
	timeFmt        string
	nower          func() time.Time
	fileNameCutter func(string) string
	formatter      func(...interface{}) (bool, string) // it has to be option too?
	lineFormatter  func(tm, level, label, caller, msg string) string
	defaultLabel   string
	commonLabel    string
	labelInfo      string
	labelError     string
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
		defaultLabel:   "",
		commonLabel:    "",
		labelInfo:      DefaultInfoLabel,
		labelError:     DefaultErrorLabel,
		output:         os.Stdout,
		callerLevel:    baseCallerLevel,
	}
	for _, o := range opt {
		o(l)
	}
	return l
}

func (l *Logger) Log(ctx context.Context, message ...interface{}) {
	tm := l.nower().Format(l.timeFmt)
	caller := l.caller()
	label := l.label(ctx)
	isError, msg := l.formatter(message...)
	level := l.labelInfo
	if isError {
		level = l.labelError
	}
	fmt.Fprintln(l.output, l.lineFormatter(tm, level, label, caller, msg))
}

func (l *Logger) label(ctx context.Context) string {
	label := l.commonLabel
	if ctx != nil {
		l, _ := ctx.Value(labelKey).(string)
		if l != "" {
			if label == "" {
				label = l
			} else {
				label += ":" + l
			}
		}
	}
	if label == "" {
		return l.defaultLabel
	}
	return label
}

func (l *Logger) caller() string {
	_, file, line, _ := runtime.Caller(l.callerLevel)
	return fmt.Sprintf("%s:%d", l.fileNameCutter(file), line)
}

// -- func

// use SetDefaultLogger to tune it.
var defaultLogger Interface = &Logger{
	timeFmt:        "2006-01-02 15:04:05",
	nower:          time.Now,
	fileNameCutter: mkLongestPrefixCutter(""),
	formatter:      defaultFormatter,
	lineFormatter:  defaultLineFomatter,
	defaultLabel:   "",
	commonLabel:    "",
	labelInfo:      DefaultInfoLabel,
	labelError:     DefaultErrorLabel,
	output:         os.Stderr,
	callerLevel:    baseCallerLevel + 1,
}

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

func WithLabelPlaceholder(s string) Option {
	return func(l *Logger) {
		l.defaultLabel = s
	}
}

func WithCommonLabel(s string) Option {
	return func(l *Logger) {
		l.commonLabel = s
	}
}

func WithLevelLabels(info, err string) Option {
	return func(l *Logger) {
		l.labelInfo = info
		l.labelError = err
	}
}

func WithCallerCutter(ctr func(p string) string) Option {
	return func(l *Logger) {
		l.fileNameCutter = ctr
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

func defaultFormatter(mm ...interface{}) (bool, string) {
	isError := false
	pp := make([]string, len(mm))
	for i, m := range mm {
		switch e := m.(type) {
		case error:
			isError = true
			if ef, ok := e.(fmt.Formatter); ok { //nolint:errorlint // here we have to check interface
				pp[i] = fmt.Sprintf("%+v", ef)
			} else {
				pp[i] = e.Error()
			}
		case string:
			pp[i] = e
		case []byte:
			if utf8.Valid(e) {
				pp[i] = string(e)
			} else {
				pp[i] = fmt.Sprintf("%v", e)
			}
		default:
			pp[i] = fmt.Sprintf("%v", e)
		}
	}
	return isError, strings.Join(pp, " ")
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
