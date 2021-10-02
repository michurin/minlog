/*
Minimalist, simple logger
*/
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
	timeFmt          string
	nower            func() time.Time
	fileNameCutter   func(string) string
	messageFormatter func(...interface{}) (bool, string) // it has to be option too?
	labelFormatter   func(interface{}) string
	lineFormatter    func(tm, level, label, caller, msg string) string
	defaultLabel     string
	commonLabel      string
	labelInfo        string
	labelError       string
	output           io.Writer
	callerLevel      int
}

type logContextKey string

var labelKey interface{} // initialized by init()

func New(opt ...Option) *Logger {
	_, file, _, _ := runtime.Caller(1)
	dirname := file[:len(file)-len(path.Base(file))] // including final separator
	l := &Logger{
		timeFmt:          "2006-01-02 15:04:05",
		nower:            time.Now,
		fileNameCutter:   mkLongestPrefixCutter(dirname),
		messageFormatter: defaultMessageFormatter,
		lineFormatter:    defaultLineFormatter,
		labelFormatter:   defaultLabelFormatter,
		defaultLabel:     "",
		commonLabel:      "",
		labelInfo:        DefaultInfoLabel,
		labelError:       DefaultErrorLabel,
		output:           os.Stdout,
		callerLevel:      baseCallerLevel,
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
	isError, msg := l.messageFormatter(message...)
	level := l.labelInfo
	if isError {
		level = l.labelError
	}
	fmt.Fprintln(l.output, l.lineFormatter(tm, level, label, caller, msg))
}

func (l *Logger) label(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	label := l.labelFormatter(ctx.Value(labelKey))
	if label == "" {
		label = l.defaultLabel
	}
	if label == "" {
		return l.commonLabel
	}
	if l.commonLabel == "" {
		return label
	}
	return l.commonLabel + ":" + label
}

func (l *Logger) caller() string {
	_, file, line, _ := runtime.Caller(l.callerLevel)
	return fmt.Sprintf("%s:%d", l.fileNameCutter(file), line)
}

// -- func

// use SetDefaultLogger to tune it.
var defaultLogger Interface

func init() {
	SetDefaultLogger(New())
	SetDefaultLabelKey(logContextKey("label"))
}

func Log(ctx context.Context, message ...interface{}) {
	defaultLogger.Log(ctx, message...)
}

// SetDefaultLogger touches global variable, it is not thread safe.
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

func WithLabelFormatter(fmtr func(interface{}) string) Option {
	return func(l *Logger) {
		l.labelFormatter = fmtr
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

func defaultMessageFormatter(mm ...interface{}) (bool, string) {
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

func defaultLineFormatter(tm, level, label, caller, msg string) string {
	a := []string{tm, level}
	if label != "" {
		a = append(a, label)
	}
	a = append(a, caller, msg)
	return strings.Join(a, " ")
}

func defaultLabelFormatter(v interface{}) string {
	label, _ := v.(string)
	return label
}

// -- context

// SetDefaultLabelKey touches global variable, it is not thread safe.
// And it obviously affects Label function. You have to set default
// label key before first Label usage.
func SetDefaultLabelKey(l interface{}) {
	labelKey = l
}

// Label sets/adds label. You are still able to use custom
// context key and custom ctx setter,
// consider SetDefaultLabelKey and WithLabelFormatter.
func Label(ctx context.Context, label string) context.Context {
	l, _ := ctx.Value(labelKey).(string)
	if l != "" {
		label = l + ":" + label
	}
	return context.WithValue(ctx, labelKey, label)
}
