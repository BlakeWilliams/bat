package stache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLexString(t *testing.T) {
	input := "<h1>Hello\nWorld</h1>"
	l := lexer{input: input, tokens: make([]lexToken, 0)}

	l.run()

	require.Len(t, l.tokens, 2)

	token := l.tokens[0]
	require.Equal(t, token.kind, lexKindText)
	require.Equal(t, token.value, "<h1>Hello\nWorld</h1>")
}

func TestLexBasicTemplate(t *testing.T) {
	input := "<h1>Hello {{name}}</h1>"
	l := lexer{input: input, tokens: make([]lexToken, 0)}

	l.run()

	require.Len(t, l.tokens, 6)

	require.Equal(t, l.tokens[0].kind, lexKindText)
	require.Equal(t, l.tokens[0].value, "<h1>Hello ")

	require.Equal(t, l.tokens[1].kind, lexKindLeftDelim)
	require.Equal(t, l.tokens[1].value, "{{")

	require.Equal(t, l.tokens[2].kind, lexKindIdentifier)
	require.Equal(t, l.tokens[2].value, "name")

	require.Equal(t, l.tokens[3].kind, lexKindRightDelim)
	require.Equal(t, l.tokens[3].value, "}}")

	require.Equal(t, l.tokens[4].kind, lexKindText)
	require.Equal(t, l.tokens[4].value, "</h1>")

	require.Equal(t, l.tokens[5].kind, lexKindEOF)
}

func TestLexDots(t *testing.T) {
	input := "{{foo.bar}}"
	l := lexer{input: input, tokens: make([]lexToken, 0)}

	l.run()

	require.Len(t, l.tokens, 6)

	require.Equal(t, l.tokens[0].kind, lexKindLeftDelim)
	require.Equal(t, l.tokens[0].value, "{{")

	require.Equal(t, l.tokens[1].kind, lexKindIdentifier)
	require.Equal(t, l.tokens[1].value, "foo")

	require.Equal(t, l.tokens[2].kind, lexKindDot)
	require.Equal(t, l.tokens[2].value, ".")

	require.Equal(t, l.tokens[3].kind, lexKindIdentifier)
	require.Equal(t, l.tokens[3].value, "bar")

	require.Equal(t, l.tokens[4].kind, lexKindRightDelim)
	require.Equal(t, l.tokens[4].value, "}}")

	require.Equal(t, l.tokens[5].kind, lexKindEOF)
}

func TestLexHash(t *testing.T) {
	input := "{{#each}}"
	l := lexer{input: input, tokens: make([]lexToken, 0)}

	l.run()

	require.Len(t, l.tokens, 5)

	require.Equal(t, l.tokens[0].kind, lexKindLeftDelim)
	require.Equal(t, l.tokens[0].value, "{{")

	require.Equal(t, l.tokens[1].kind, lexKindHash)
	require.Equal(t, l.tokens[1].value, "#")

	require.Equal(t, l.tokens[2].kind, lexKindIdentifier)
	require.Equal(t, l.tokens[2].value, "each")

	require.Equal(t, l.tokens[3].kind, lexKindRightDelim)
	require.Equal(t, l.tokens[3].value, "}}")

	require.Equal(t, l.tokens[4].kind, lexKindEOF)
}
