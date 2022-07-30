package stache

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type lexToken struct {
	kind  lexKind
	value string
}

type lexer struct {
	input  string
	start  int
	pos    int
	tokens []lexToken
}

type lexKind int

const (
	lexKindError lexKind = iota
	// Represents raw text in a template
	lexKindText
	lexKindEOF
	lexKindLeftDelim
	lexKindRightDelim
	lexKindIdentifier
	lexKindDot
	lexKindHash
)

type stateFn func(*lexer) stateFn

const eof = -1

const (
	leftDelim  = "{{"
	rightDelim = "}}"
)

func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
}

func (l *lexer) emit(kind lexKind) {
	l.tokens = append(l.tokens, lexToken{kind: kind, value: l.input[l.start:l.pos]})
	l.start = l.pos
	l.pos = l.start
}

func (l *lexer) emitError(content string) {
	l.tokens = append(l.tokens, lexToken{kind: lexKindError, value: content})
}

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		return eof
	}

	r, width := utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += width

	return r
}

func (l *lexer) backup() {
	_, width := utf8.DecodeLastRuneInString(l.input[:l.pos])
	l.pos -= width
}

func (l *lexer) peek() rune {
	pos := l.pos
	r := l.next()
	l.pos = pos

	return r
}

func lexText(l *lexer) stateFn {
	if index := strings.Index(l.input[l.start:], leftDelim); index >= 0 {
		if index > 0 {
			l.pos = index
			l.emit(lexKindText)
		}

		return lexLeftDelim
	}

	// If there's remaining text, emit it
	if l.start != len(l.input) {
		l.pos = len(l.input)
		l.emit(lexKindText)
	}

	l.emit(lexKindEOF)

	return nil
}

func lexLeftDelim(l *lexer) stateFn {
	l.pos += len(leftDelim)
	l.emit(lexKindLeftDelim)

	return lexAction
}

func lexAction(l *lexer) stateFn {
	r := l.peek()
	switch {
	case r == '}':
		return lexRightDelim
	case r == '.':
		l.next()
		l.emit(lexKindDot)
		return lexAction
	case r == '#':
		l.next()
		l.emit(lexKindHash)
		return lexAction
	case unicode.IsLetter(r):
		return lexIdentifier
	}
	return nil
}

func lexRightDelim(l *lexer) stateFn {
	if !strings.HasPrefix(l.input[l.pos:], rightDelim) {
		l.emitError(fmt.Sprintf("expected }}, got %s", l.input[l.pos:l.pos+2]))
		return nil
	}

	l.pos += len(rightDelim)
	l.emit(lexKindRightDelim)

	return lexText
}

func lexIdentifier(l *lexer) stateFn {
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

	l.emit(lexKindIdentifier)

	return lexAction
}
