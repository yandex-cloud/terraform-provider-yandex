package log

import (
	"context"
	"fmt"
	"io"

	"github.com/jonboulle/clockwork"

	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
)

const (
	dateLayout = "2006-01-02 15:04:05.000"
)

type Logger interface {
	// Log logs the message with specified options and fields.
	// Implementations must not in any way use slice of fields after Log returns.
	Log(ctx context.Context, msg string, fields ...Field)
}

var _ Logger = (*defaultLogger)(nil)

type simpleLoggerOption interface {
	applySimpleOption(l *defaultLogger)
}

func Default(w io.Writer, opts ...simpleLoggerOption) *defaultLogger {
	l := &defaultLogger{
		coloring: false,
		minLevel: INFO,
		clock:    clockwork.NewRealClock(),
		w:        w,
	}
	for _, opt := range opts {
		if opt != nil {
			opt.applySimpleOption(l)
		}
	}

	return l
}

type defaultLogger struct {
	coloring bool
	logQuery bool
	minLevel Level
	clock    clockwork.Clock
	w        io.Writer
}

func (l *defaultLogger) format(namespace []string, msg string, logLevel Level) string {
	b := xstring.Buffer()
	defer b.Free()
	if l.coloring {
		b.WriteString(logLevel.Color())
	}
	b.WriteString(l.clock.Now().Format(dateLayout))
	b.WriteByte(' ')
	lvl := logLevel.String()
	if l.coloring {
		b.WriteString(colorReset)
		b.WriteString(logLevel.BoldColor())
	}
	b.WriteString(lvl)
	if l.coloring {
		b.WriteString(colorReset)
		b.WriteString(logLevel.Color())
	}
	b.WriteString(" '")
	for i, name := range namespace {
		if i != 0 {
			b.WriteByte('.')
		}
		b.WriteString(name)
	}
	b.WriteString("' => ")
	b.WriteString(msg)
	if l.coloring {
		b.WriteString(colorReset)
	}

	return b.String()
}

func (l *defaultLogger) Log(ctx context.Context, msg string, fields ...Field) {
	lvl := LevelFromContext(ctx)
	if lvl < l.minLevel {
		return
	}

	_, _ = io.WriteString(l.w, l.format(
		NamesFromContext(ctx),
		l.appendFields(msg, fields...),
		lvl,
	)+"\n")
}

type wrapper struct {
	logQuery bool
	logger   Logger
}

func wrapLogger(l Logger, opts ...Option) *wrapper {
	ll := &wrapper{
		logger: l,
	}
	for _, opt := range opts {
		if opt != nil {
			opt.applyHolderOption(ll)
		}
	}

	return ll
}

func (l *defaultLogger) appendFields(msg string, fields ...Field) string {
	if len(fields) == 0 {
		return msg
	}
	b := xstring.Buffer()
	defer b.Free()
	b.WriteString(msg)
	b.WriteString(" {")
	for i := range fields {
		if i != 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(b, `%q:%q`, fields[i].Key(), fields[i].String())
	}
	b.WriteByte('}')

	return b.String()
}

func (l *wrapper) Log(ctx context.Context, msg string, fields ...Field) {
	l.logger.Log(ctx, msg, fields...)
}
