package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type (
	Lexer struct {
		input     string
		start     int
		pos       int
		Tokens    []Token
		Line      int
		StartLine int
	}

	Kind int

	stateFn func(*Lexer) stateFn
)

const eof = -1

const (
	leftDelim  = "{{"
	rightDelim = "}}"
)

func Lex(input string) *Lexer {
	l := &Lexer{input: input, Tokens: make([]Token, 0), StartLine: 1, Line: 1}
	l.run()

	return l
}

func (l *Lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
}

func (l *Lexer) currentText() string {
	return l.input[l.start:l.pos]
}

func (l *Lexer) emit(kind Kind) {
	token := Token{
		Kind:      kind,
		Value:     l.input[l.start:l.pos],
		StartLine: l.StartLine,
		EndLine:   l.Line,
	}

	l.StartLine = l.Line
	l.Tokens = append(l.Tokens, token)
	l.start = l.pos
	l.pos = l.start
}

func (l *Lexer) emitError(content string) {
	l.Tokens = append(l.Tokens, Token{Kind: KindError, Value: content})
}

func (l *Lexer) next() rune {
	if l.pos >= len(l.input) {
		return eof
	}

	r, width := utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += width

	if r == '\n' {
		l.Line++
	}

	return r
}

func (l *Lexer) backup() {
	r, width := utf8.DecodeLastRuneInString(l.input[:l.pos])

	if r == '\n' {
		l.Line -= 1
	}

	l.pos -= width
}

func (l *Lexer) peek() rune {
	r := l.next()
	l.backup()

	return r
}

func lexText(l *Lexer) stateFn {
	if index := strings.Index(l.input[l.start:], leftDelim); index >= 0 {
		if index > 0 {
			l.pos = l.start + index

			l.Line += strings.Count(l.input[l.start:l.pos], "\n")
			l.emit(KindText)
		}

		return lexLeftDelim
	}

	// If there's remaining text, emit it
	if l.start != len(l.input) {
		l.pos = len(l.input)
		l.emit(KindText)
	}

	l.emit(KindEOF)

	return nil
}

func lexLeftDelim(l *Lexer) stateFn {
	l.pos += len(leftDelim)
	l.emit(KindLeftDelim)

	return lexAction
}

func lexAction(l *Lexer) stateFn {
	r := l.peek()
	switch {
	case r == '}':
		return lexRightDelim
	case r == '.':
		l.next()
		l.emit(KindDot)
		return lexAction
	case r == '#':
		l.next()
		l.emit(KindHash)
		return lexAction
	case r == '=':
		l.next()
		l.emit(KindEqual)
		return lexAction
	case r == '!':
		l.next()
		l.emit(KindBang)
		return lexAction
	case unicode.IsSpace(r):
		return lexSpace
	case unicode.IsLetter(r):
		return lexIdentifier
	}
	return nil
}

func lexRightDelim(l *Lexer) stateFn {
	if !strings.HasPrefix(l.input[l.pos:], rightDelim) {
		l.emitError(fmt.Sprintf("expected }}, got %s", l.input[l.pos:l.pos+2]))
		return nil
	}

	l.pos += len(rightDelim)
	l.emit(KindRightDelim)

	return lexText
}

func lexIdentifier(l *Lexer) stateFn {
	for {
		r := l.next()

		if r == eof {
			break
		}

		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			l.backup()
			break
		}
	}

	switch l.currentText() {
	case "if":
		l.emit(KindIf)
	case "else":
		l.emit(KindElse)
	case "nil":
		l.emit(KindNil)
	case "end":
		l.emit(KindEnd)
	case "true":
		l.emit(KindTrue)
	case "false":
		l.emit(KindFalse)
	default:
		l.emit(KindIdentifier)
	}

	return lexAction
}

func lexSpace(l *Lexer) stateFn {
	for {
		r := l.next()

		if r == eof {
			break
		}

		if !unicode.IsSpace(r) {
			l.backup()
			break
		}
	}

	l.emit(KindSpace)

	return lexAction
}
