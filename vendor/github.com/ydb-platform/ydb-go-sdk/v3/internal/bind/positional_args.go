package bind

import (
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
)

type PositionalArgs struct{}

func (m PositionalArgs) blockID() blockID {
	return blockYQL
}

func (m PositionalArgs) ToYdb(sql string, args ...any) (
	yql string, newArgs []any, err error,
) {
	l := &sqlLexer{
		src:        sql,
		stateFn:    positionalArgsStateFn,
		rawStateFn: positionalArgsStateFn,
	}

	for l.stateFn != nil {
		l.stateFn = l.stateFn(l)
	}

	var (
		buffer   = xstring.Buffer()
		position = 0
		param    table.ParameterOption
	)
	defer buffer.Free()

	for _, p := range l.parts {
		switch p := p.(type) {
		case string:
			buffer.WriteString(p)
		case positionalArg:
			if position > len(args)-1 {
				return "", nil, xerrors.WithStackTrace(
					fmt.Errorf("%w: position %d, len(args) = %d", ErrInconsistentArgs, position, len(args)),
				)
			}
			paramName := "$p" + strconv.Itoa(position)
			param, err = toYdbParam(paramName, args[position])
			if err != nil {
				return "", nil, xerrors.WithStackTrace(err)
			}
			newArgs = append(newArgs, param)
			buffer.WriteString(paramName)
			position++
		}
	}

	if len(args) != position {
		return "", nil, xerrors.WithStackTrace(
			fmt.Errorf("%w: (positional args %d, query args %d)", ErrInconsistentArgs, position, len(args)),
		)
	}

	if position > 0 {
		const prefix = "-- origin query with positional args replacement\n"

		return prefix + buffer.String(), newArgs, nil
	}

	return buffer.String(), newArgs, nil
}

func positionalArgsStateFn(l *sqlLexer) stateFn {
	for {
		r, width := utf8.DecodeRuneInString(l.src[l.pos:])
		l.pos += width

		switch r {
		case '`':
			return backtickState
		case '\'':
			return singleQuoteState
		case '"':
			return doubleQuoteState
		case '?':
			l.parts = append(l.parts, l.src[l.start:l.pos-1], positionalArg{})
			l.start = l.pos
		case '-':
			nextRune, width := utf8.DecodeRuneInString(l.src[l.pos:])
			if nextRune == '-' {
				l.pos += width

				return oneLineCommentState
			}
		case '/':
			nextRune, width := utf8.DecodeRuneInString(l.src[l.pos:])
			if nextRune == '*' {
				l.pos += width

				return multilineCommentState
			}
		case utf8.RuneError:
			if l.pos-l.start > 0 {
				l.parts = append(l.parts, l.src[l.start:l.pos])
				l.start = l.pos
			}

			return nil
		}
	}
}
