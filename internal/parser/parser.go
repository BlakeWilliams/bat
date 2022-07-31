package parser

import (
	"fmt"
	"strings"

	"github.com/blakewilliams/stache/internal/lexer"
)

// Represents a node in the template AST (abstract syntax tree).
type Node struct {
	Kind     string
	Children []*Node
	Value    string
}

type parser struct {
	lexer *lexer.Lexer
	Root  *Node
	pos   int
}

const (
	KindRoot       = "root"
	KindText       = "text"
	KindStatement  = "statement"
	KindAccess     = "access"
	KindIdentifier = "identifier"
)

func (n *Node) String() string {
	out := fmt.Sprintf("(%s", n.Kind)

	if n.Value != "" {
		out += fmt.Sprintf(" `%s`", n.Value)
	}

	if len(n.Children) > 0 {
		out += "\n"

		for i, child := range n.Children {
			str := child.String()
			str = "   " + strings.Join(strings.Split(str, "\n"), "\n   ")
			out += str

			if i < len(n.Children)-1 {
				out += "\n"
			}
		}
	}

	out += ")"

	return out
}

func (p *parser) peek() lexer.Token {
	return p.lexer.Tokens[p.pos+1]
}

func (p *parser) next() lexer.Token {
	p.pos++
	return p.lexer.Tokens[p.pos]
}

func (p *parser) skipWhitespace() {
	for {
		if p.peek().Kind != lexer.KindSpace {
			break
		}
		p.next()
	}
}

func Parse(l *lexer.Lexer) *Node {
	p := &parser{
		lexer: l,
		Root:  &Node{Kind: KindRoot},
		pos:   -1,
	}

	p.Root.Children = parseMany(p)

	return p.Root
}

func parseMany(p *parser) []*Node {
	nodes := make([]*Node, 0)

	for {
		switch p.peek().Kind {
		case lexer.KindEOF:
			return nodes
		case lexer.KindText:
			node := &Node{Kind: KindText, Value: p.next().Value}
			nodes = append(nodes, node)
		case lexer.KindLeftDelim:
			node := &Node{Kind: KindStatement}
			p.next()

			node.Children = parseStatement(p)
			nodes = append(nodes, node)
		default:
			panic(fmt.Sprintf("unsupported token %v", p.peek()))
		}
	}
}

func parseStatement(p *parser) []*Node {
	nodes := make([]*Node, 0)
	p.skipWhitespace()

	for {
		switch p.peek().Kind {
		case lexer.KindRightDelim:
			p.next()
			return nodes
		case lexer.KindEOF:
			panic("unexpected EOF")
		case lexer.KindIdentifier:
			node := parseExpression(p)
			nodes = append(nodes, node)
		case lexer.KindSpace:
			p.skipWhitespace()
		default:
			panic(fmt.Sprintf("unexpected token %v", p.peek()))
		}
	}
}

func parseExpression(p *parser) *Node {
	identifierToken := p.next()
	identifierNode := &Node{Kind: KindIdentifier, Value: identifierToken.Value}

	if p.peek().Kind == lexer.KindDot {
		node := identifierNode

		for p.peek().Kind == lexer.KindDot {
			p.expect(lexer.KindDot)

			identifier := p.expect(lexer.KindIdentifier)
			identifierNode := &Node{Kind: KindIdentifier, Value: identifier.Value}

			newNode := &Node{Kind: KindAccess, Children: []*Node{node, identifierNode}}
			node = newNode
		}

		return node
	}

	return identifierNode
}

func (p *parser) expect(kind lexer.Kind) lexer.Token {
	n := p.next()

	if n.Kind != kind {
		panic(fmt.Sprintf("unexpected token %v, expected %s", n, kind))
	}

	return n
}
