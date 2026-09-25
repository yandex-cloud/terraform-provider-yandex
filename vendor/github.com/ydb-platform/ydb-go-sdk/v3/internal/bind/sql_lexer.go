package bind

import "unicode/utf8"

type sqlLexer struct {
	src        string
	start      int
	pos        int
	nested     int // multiline comment nesting level.
	stateFn    stateFn
	rawStateFn stateFn
	parts      []any
}

type (
	positionalArg struct{}
	numericArg    int

	stateFn func(*sqlLexer) stateFn
)

func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isNumber(r rune) bool {
	return r >= '0' && r <= '9'
}

func backtickState(l *sqlLexer) stateFn {
	for {
		r, width := utf8.DecodeRuneInString(l.src[l.pos:])
		l.pos += width

		switch r {
		case '`':
			nextRune, width := utf8.DecodeRuneInString(l.src[l.pos:])
			if nextRune != '`' {
				return l.rawStateFn
			}
			l.pos += width
		case utf8.RuneError:
			if l.pos-l.start > 0 {
				l.parts = append(l.parts, l.src[l.start:l.pos])
				l.start = l.pos
			}

			return nil
		}
	}
}

func singleQuoteState(l *sqlLexer) stateFn {
	for {
		r, width := utf8.DecodeRuneInString(l.src[l.pos:])
		l.pos += width

		switch r {
		case '\'':
			nextRune, width := utf8.DecodeRuneInString(l.src[l.pos:])
			if nextRune != '\'' {
				return l.rawStateFn
			}
			l.pos += width
		case utf8.RuneError:
			if l.pos-l.start > 0 {
				l.parts = append(l.parts, l.src[l.start:l.pos])
				l.start = l.pos
			}

			return nil
		}
	}
}

func doubleQuoteState(l *sqlLexer) stateFn {
	for {
		r, width := utf8.DecodeRuneInString(l.src[l.pos:])
		l.pos += width

		switch r {
		case '"':
			nextRune, width := utf8.DecodeRuneInString(l.src[l.pos:])
			if nextRune != '"' {
				return l.rawStateFn
			}
			l.pos += width
		case utf8.RuneError:
			if l.pos-l.start > 0 {
				l.parts = append(l.parts, l.src[l.start:l.pos])
				l.start = l.pos
			}

			return nil
		}
	}
}

func oneLineCommentState(l *sqlLexer) stateFn {
	for {
		r, width := utf8.DecodeRuneInString(l.src[l.pos:])
		l.pos += width

		switch r {
		case '\\':
			_, width = utf8.DecodeRuneInString(l.src[l.pos:])
			l.pos += width
		case '\n', '\r':
			return l.rawStateFn
		case utf8.RuneError:
			if l.pos-l.start > 0 {
				l.parts = append(l.parts, l.src[l.start:l.pos])
				l.start = l.pos
			}

			return nil
		}
	}
}

func multilineCommentState(l *sqlLexer) stateFn {
	for {
		r, width := utf8.DecodeRuneInString(l.src[l.pos:])
		l.pos += width

		switch r {
		case '/':
			nextRune, width := utf8.DecodeRuneInString(l.src[l.pos:])
			if nextRune == '*' {
				l.pos += width
				l.nested++
			}
		case '*':
			nextRune, width := utf8.DecodeRuneInString(l.src[l.pos:])
			if nextRune != '/' {
				continue
			}

			l.pos += width
			if l.nested == 0 {
				return l.rawStateFn
			}
			l.nested--

		case utf8.RuneError:
			if l.pos-l.start > 0 {
				l.parts = append(l.parts, l.src[l.start:l.pos])
				l.start = l.pos
			}

			return nil
		}
	}
}
