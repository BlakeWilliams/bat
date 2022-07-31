package parser

import (
	"testing"

	"github.com/blakewilliams/bat/internal/lexer"
	"github.com/stretchr/testify/require"
)

type testNode struct {
	Kind     string
	Children []testNode
	Value    string
}

func TestParse(t *testing.T) {
	l := lexer.Lex("<h1>Hello world</h1>")
	result := Parse(l)

	require.Len(t, result.Children, 1)

	node := result.Children[0]

	require.Equal(t, node.Kind, KindText)
	require.Equal(t, node.Value, "<h1>Hello world</h1>")
}

func TestParseDelim(t *testing.T) {
	l := lexer.Lex("<h1>Hello {{name}}</h1>")
	result := Parse(l)

	require.Len(t, result.Children, 3)

	node := result.Children[0]
	require.Equal(t, node.Kind, KindText)
	require.Equal(t, node.Value, "<h1>Hello ")

	node = result.Children[1]
	require.Equal(t, node.Kind, KindStatement)

	require.Len(t, node.Children, 1)
	require.Equal(t, node.Children[0].Kind, KindIdentifier)
	require.Equal(t, node.Children[0].Value, "name")

	node = result.Children[2]
	require.Equal(t, node.Kind, KindText)
	require.Equal(t, node.Value, "</h1>")
}

func TestParseDots(t *testing.T) {
	l := lexer.Lex("<h1>Hello {{  foo.bar.baz   }}</h1>")
	result := Parse(l)

	expected := n(KindRoot, "", []*Node{
		n(KindText, "<h1>Hello ", nil),
		n(KindStatement, "", []*Node{
			n(KindAccess, "", []*Node{
				n(KindAccess, "", []*Node{
					n(KindIdentifier, "foo", nil),
					n(KindIdentifier, "bar", nil),
				}),
				n(KindIdentifier, "baz", nil),
			}),
		}),
		n(KindText, "</h1>", nil),
	})

	require.Equal(t, expected.String(), result.String())
}

func TestParse_If(t *testing.T) {
	l := lexer.Lex("{{if name != nil}}Hello!{{else}}Goodbye!{{end}}")
	result := Parse(l)

	expected := n(KindRoot, "", []*Node{
		n(KindStatement, "", []*Node{
			n(KindIf, "", []*Node{
				n(KindInfix, "", []*Node{
					n(KindIdentifier, "name", nil),
					n(KindOperator, "!=", nil),
					n(KindNil, "", nil),
				}),
				n(KindText, "Hello!", nil),
				n(KindText, "Goodbye!", nil),
			}),
		}),
	})

	require.Equal(t, expected.String(), result.String())
}

func n(kind string, value string, children []*Node) *Node {
	return &Node{Kind: kind, Value: value, Children: children}
}
