package lexer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLexString(t *testing.T) {
	input := "<h1>Hello\nWorld</h1>"
	l := Lexer{input: input, Tokens: make([]Token, 0)}

	l.run()

	require.Len(t, l.Tokens, 2)

	token := l.Tokens[0]
	require.Equal(t, token.Kind, KindText)
	require.Equal(t, token.Value, "<h1>Hello\nWorld</h1>")
}

func TestLexBasicTemplate(t *testing.T) {
	input := "<h1>Hello {{name}}</h1>"
	l := Lexer{input: input, Tokens: make([]Token, 0)}

	l.run()

	require.Len(t, l.Tokens, 6)

	require.Equal(t, l.Tokens[0].Kind, KindText)
	require.Equal(t, l.Tokens[0].Value, "<h1>Hello ")

	require.Equal(t, l.Tokens[1].Kind, KindLeftDelim)
	require.Equal(t, l.Tokens[1].Value, "{{")

	require.Equal(t, l.Tokens[2].Kind, KindIdentifier)
	require.Equal(t, l.Tokens[2].Value, "name")

	require.Equal(t, l.Tokens[3].Kind, KindRightDelim)
	require.Equal(t, l.Tokens[3].Value, "}}")

	require.Equal(t, l.Tokens[4].Kind, KindText)
	require.Equal(t, l.Tokens[4].Value, "</h1>")

	require.Equal(t, l.Tokens[5].Kind, KindEOF)
}

func TestLexDots(t *testing.T) {
	input := "{{foo.bar}}"
	l := Lexer{input: input, Tokens: make([]Token, 0)}

	l.run()

	require.Len(t, l.Tokens, 6)

	require.Equal(t, l.Tokens[0].Kind, KindLeftDelim)
	require.Equal(t, l.Tokens[0].Value, "{{")

	require.Equal(t, l.Tokens[1].Kind, KindIdentifier)
	require.Equal(t, l.Tokens[1].Value, "foo")

	require.Equal(t, l.Tokens[2].Kind, KindDot)
	require.Equal(t, l.Tokens[2].Value, ".")

	require.Equal(t, l.Tokens[3].Kind, KindIdentifier)
	require.Equal(t, l.Tokens[3].Value, "bar")

	require.Equal(t, l.Tokens[4].Kind, KindRightDelim)
	require.Equal(t, l.Tokens[4].Value, "}}")

	require.Equal(t, l.Tokens[5].Kind, KindEOF)
}

func TestLexMultipleStatements(t *testing.T) {
	input := "{{foo.bar}} {{bar.baz}}"
	l := Lexer{input: input, Tokens: make([]Token, 0)}

	l.run()
	fmt.Println(l.Tokens)

	require.Len(t, l.Tokens, 12)

	require.Equal(t, l.Tokens[0].Kind, KindLeftDelim)
	require.Equal(t, l.Tokens[0].Value, "{{")

	require.Equal(t, l.Tokens[1].Kind, KindIdentifier)
	require.Equal(t, l.Tokens[1].Value, "foo")

	require.Equal(t, l.Tokens[2].Kind, KindDot)
	require.Equal(t, l.Tokens[2].Value, ".")

	require.Equal(t, l.Tokens[3].Kind, KindIdentifier)
	require.Equal(t, l.Tokens[3].Value, "bar")

	require.Equal(t, l.Tokens[4].Kind, KindRightDelim)
	require.Equal(t, l.Tokens[4].Value, "}}")

	require.Equal(t, l.Tokens[6].Kind, KindLeftDelim)
	require.Equal(t, l.Tokens[6].Value, "{{")

	require.Equal(t, l.Tokens[7].Kind, KindIdentifier)
	require.Equal(t, l.Tokens[7].Value, "bar")

	require.Equal(t, l.Tokens[8].Kind, KindDot)
	require.Equal(t, l.Tokens[8].Value, ".")

	require.Equal(t, l.Tokens[9].Kind, KindIdentifier)
	require.Equal(t, l.Tokens[9].Value, "baz")

	require.Equal(t, l.Tokens[10].Kind, KindRightDelim)
	require.Equal(t, l.Tokens[10].Value, "}}")

	require.Equal(t, l.Tokens[11].Kind, KindEOF)
}

func TestLexHash(t *testing.T) {
	input := "{{#each}}"
	l := Lexer{input: input, Tokens: make([]Token, 0)}

	l.run()

	require.Len(t, l.Tokens, 5)

	require.Equal(t, l.Tokens[0].Kind, KindLeftDelim)
	require.Equal(t, l.Tokens[0].Value, "{{")

	require.Equal(t, l.Tokens[1].Kind, KindHash)
	require.Equal(t, l.Tokens[1].Value, "#")

	require.Equal(t, l.Tokens[2].Kind, KindIdentifier)
	require.Equal(t, l.Tokens[2].Value, "each")

	require.Equal(t, l.Tokens[3].Kind, KindRightDelim)
	require.Equal(t, l.Tokens[3].Value, "}}")

	require.Equal(t, l.Tokens[4].Kind, KindEOF)
}

func TestLexSpaces(t *testing.T) {
	input := "{{   #each   }}"
	l := Lexer{input: input, Tokens: make([]Token, 0)}

	l.run()

	require.Len(t, l.Tokens, 7)

	require.Equal(t, l.Tokens[0].Kind, KindLeftDelim)
	require.Equal(t, l.Tokens[0].Value, "{{")

	require.Equal(t, l.Tokens[1].Kind, KindSpace)
	require.Equal(t, l.Tokens[1].Value, "   ")

	require.Equal(t, l.Tokens[2].Kind, KindHash)
	require.Equal(t, l.Tokens[2].Value, "#")

	require.Equal(t, l.Tokens[3].Kind, KindIdentifier)
	require.Equal(t, l.Tokens[3].Value, "each")

	require.Equal(t, l.Tokens[4].Kind, KindSpace)
	require.Equal(t, l.Tokens[4].Value, "   ")

	require.Equal(t, l.Tokens[5].Kind, KindRightDelim)
	require.Equal(t, l.Tokens[5].Value, "}}")

	require.Equal(t, l.Tokens[6].Kind, KindEOF)
}
