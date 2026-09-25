package bind

import (
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/params"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
)

type NumericArgs struct{}

func (m NumericArgs) blockID() blockID {
	return blockYQL
}

func (m NumericArgs) ToYdb(sql string, args ...any) (yql string, newArgs []any, err error) {
	l := &sqlLexer{
		src:        sql,
		stateFn:    numericArgsStateFn,
		rawStateFn: numericArgsStateFn,
	}

	for l.stateFn != nil {
		l.stateFn = l.stateFn(l)
	}

	buffer := xstring.Buffer()
	defer buffer.Free()

	if len(args) > 0 {
		parameters, err := parsePositionalParameters(args)
		if err != nil {
			return "", nil, err
		}
		newArgs = make([]any, len(parameters))
		for i, param := range parameters {
			newArgs[i] = param
		}
	}

	for _, p := range l.parts {
		switch p := p.(type) {
		case string:
			buffer.WriteString(p)
		case numericArg:
			if p == 0 {
				return "", nil, xerrors.WithStackTrace(ErrUnexpectedNumericArgZero)
			}
			if int(p) > len(args) {
				return "", nil, xerrors.WithStackTrace(
					fmt.Errorf("%w: $%d, len(args) = %d", ErrInconsistentArgs, p, len(args)),
				)
			}
			paramIndex := int(p - 1)
			val, ok := newArgs[paramIndex].(table.ParameterOption)
			if !ok {
				panic(fmt.Sprintf("unsupported type conversion from %T to table.ParameterOption", val))
			}
			buffer.WriteString(val.Name())
		}
	}

	yql = buffer.String()
	if len(newArgs) > 0 {
		yql = "-- origin query with numeric args replacement\n" + yql
	}

	return yql, newArgs, nil
}

func numericArgsStateFn(l *sqlLexer) stateFn {
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
		case '$':
			nextRune, _ := utf8.DecodeRuneInString(l.src[l.pos:])
			if isNumber(nextRune) {
				if l.pos-l.start > 0 {
					l.parts = append(l.parts, l.src[l.start:l.pos-width])
				}
				l.start = l.pos

				return numericArgState
			}
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

func parsePositionalParameters(args []any) ([]*params.Parameter, error) {
	newArgs := make([]*params.Parameter, len(args))
	for i, arg := range args {
		paramName := fmt.Sprintf("$p%d", i)
		param, err := toYdbParam(paramName, arg)
		if err != nil {
			return nil, err
		}
		newArgs[i] = param
	}

	return newArgs, nil
}

func numericArgState(l *sqlLexer) stateFn {
	numbers := ""
	defer func() {
		if len(numbers) > 0 {
			i, err := strconv.Atoi(numbers)
			if err != nil {
				panic(err)
			}
			l.parts = append(l.parts, numericArg(i))
			l.start = l.pos
		} else {
			l.parts = append(l.parts, l.src[l.start-1:l.pos])
			l.start = l.pos
		}
	}()
	for {
		r, width := utf8.DecodeRuneInString(l.src[l.pos:])
		l.pos += width

		switch {
		case isNumber(r):
			numbers += string(r)
		case isLetter(r):
			numbers = ""

			return l.rawStateFn
		default:
			l.pos -= width

			return l.rawStateFn
		}
	}
}
