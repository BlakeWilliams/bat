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

func TestLex_Pos(t *testing.T) {
	input := "<h1>\nHello\n{{\nname\n}}</h1>"
	l := Lex(input)

	l.run()

	require.Equal(t, l.Tokens[0].Kind, KindText)
	require.Equal(t, l.Tokens[0].Value, "<h1>\nHello\n")
	require.Equal(t, l.Tokens[0].StartLine, 1)
	require.Equal(t, l.Tokens[0].EndLine, 3)

	require.Equal(t, l.Tokens[1].Kind, KindLeftDelim)
	require.Equal(t, l.Tokens[1].Value, "{{")
	require.Equal(t, l.Tokens[1].StartLine, 3)
	require.Equal(t, l.Tokens[1].EndLine, 3)

	fmt.Println(input)
	require.Equal(t, l.Tokens[3].Kind, KindIdentifier)
	require.Equal(t, l.Tokens[3].Value, "name")
	require.Equal(t, l.Tokens[3].StartLine, 4)
	require.Equal(t, l.Tokens[3].EndLine, 4)

	require.Equal(t, l.Tokens[5].Kind, KindRightDelim)
	require.Equal(t, l.Tokens[5].Value, "}}")
	require.Equal(t, l.Tokens[5].StartLine, 5)
	require.Equal(t, l.Tokens[5].EndLine, 5)

	require.Equal(t, l.Tokens[6].Kind, KindText)
	require.Equal(t, l.Tokens[6].Value, "</h1>")
	require.Equal(t, l.Tokens[6].StartLine, 5)
	require.Equal(t, l.Tokens[6].EndLine, 5)

	require.Equal(t, l.Tokens[7].Kind, KindEOF)
}

func TestLex_If(t *testing.T) {
	input := "{{if foo != nil}}1{{else}}2{{end}}"
	l := Lexer{input: input, Tokens: make([]Token, 0)}

	l.run()
	fmt.Println(l)

	require.Equal(t, l.Tokens[0].Kind, KindLeftDelim)
	require.Equal(t, l.Tokens[0].Value, "{{")

	require.Equal(t, l.Tokens[1].Kind, KindIf)
	require.Equal(t, l.Tokens[1].Value, "if")

	require.Equal(t, l.Tokens[2].Kind, KindSpace)
	require.Equal(t, l.Tokens[3].Kind, KindIdentifier)
	require.Equal(t, l.Tokens[4].Kind, KindSpace)

	require.Equal(t, l.Tokens[5].Kind, KindBang)
	require.Equal(t, l.Tokens[5].Value, "!")

	require.Equal(t, l.Tokens[6].Kind, KindEqual)
	require.Equal(t, l.Tokens[6].Value, "=")

	require.Equal(t, l.Tokens[7].Kind, KindSpace)
	require.Equal(t, l.Tokens[8].Kind, KindNil)

	require.Equal(t, l.Tokens[9].Kind, KindRightDelim)
	require.Equal(t, l.Tokens[10].Kind, KindText)
	require.Equal(t, l.Tokens[11].Kind, KindLeftDelim)
	require.Equal(t, l.Tokens[12].Kind, KindElse)
	require.Equal(t, l.Tokens[13].Kind, KindRightDelim)
	require.Equal(t, l.Tokens[14].Kind, KindText)
	require.Equal(t, l.Tokens[15].Kind, KindLeftDelim)
	require.Equal(t, l.Tokens[16].Kind, KindEnd)
	require.Equal(t, l.Tokens[17].Kind, KindRightDelim)
}
