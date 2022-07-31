package parser

import (
	"fmt"
	"strings"

	"github.com/blakewilliams/bat/internal/lexer"
)

// Represents a node in the template AST (abstract syntax tree).
type Node struct {
	Kind      string
	Children  []*Node
	Value     string
	StartLine int
	EndLine   int
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
	KindIf         = "if"
	KindInfix      = "infix"
	KindOperator   = "operator"
	KindNil        = "nil"
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

func (p *parser) peekn(n int) lexer.Token {
	return p.lexer.Tokens[p.pos+n]
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
			token := p.next()
			node := &Node{Kind: KindText, Value: token.Value, StartLine: token.StartLine, EndLine: token.EndLine}
			nodes = append(nodes, node)
		case lexer.KindLeftDelim:
			p.skipWhitespace()
			token := p.next()
			// else and end signify the end of the current statement, so exit
			switch p.peek().Kind {
			case lexer.KindElse:
				return nodes
			case lexer.KindEnd:
				return nodes
			}

			// parse everything between {{ and }}
			node := &Node{Kind: KindStatement, StartLine: token.StartLine, EndLine: token.EndLine}
			node.Children = parseStatement(p)
			nodes = append(nodes, node)
		case lexer.KindElse:
			return nodes
		case lexer.KindEnd:
			return nodes
		default:
			panic(fmt.Sprintf("unsupported token %v", p.peek()))
		}
	}
}

// Statements represent everything in a `{{...}}` block.
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
		case lexer.KindNil:
			token := p.next()
			node := &Node{Kind: KindNil, StartLine: token.StartLine, EndLine: token.EndLine}
			nodes = append(nodes, node)
		case lexer.KindSpace:
			p.skipWhitespace()
		case lexer.KindIf:
			node := parseIf(p)
			nodes = append(nodes, node)
		default:
			panic(fmt.Sprintf("unexpected token %v", p.peek()))
		}
	}
}

// parses expressions, like:
// foo.bar.baz
// foo != nil
func parseExpression(p *parser) *Node {
	identifierToken := p.next()
	if identifierToken.Kind == lexer.KindNil {
		return &Node{
			Kind:      KindNil,
			StartLine: identifierToken.StartLine,
			EndLine:   identifierToken.EndLine,
		}
	}
	identifierNode := &Node{
		Kind:      KindIdentifier,
		Value:     identifierToken.Value,
		StartLine: identifierToken.StartLine,
		EndLine:   identifierToken.EndLine,
	}

	if p.peek().Kind == lexer.KindDot {
		node := identifierNode

		for p.peek().Kind == lexer.KindDot {
			p.expect(lexer.KindDot)

			identifier := p.expect(lexer.KindIdentifier)
			identifierNode := &Node{
				Kind:      KindIdentifier,
				Value:     identifier.Value,
				StartLine: identifier.StartLine,
				EndLine:   identifier.EndLine,
			}

			newNode := &Node{
				Kind:      KindAccess,
				Children:  []*Node{node, identifierNode},
				StartLine: identifier.StartLine,
				EndLine:   identifier.EndLine,
			}
			node = newNode
		}

		return node
	}

	p.skipWhitespace()

	if p.peek().Kind == lexer.KindBang || p.peek().Kind == lexer.KindEqual {
		operator := parseOperator(p)
		p.skipWhitespace()

		node := &Node{
			Kind:      KindInfix,
			Children:  []*Node{},
			StartLine: identifierToken.StartLine,
			EndLine:   p.peek().EndLine,
		}

		node.Children = append(node.Children, identifierNode)
		node.Children = append(node.Children, operator)
		node.Children = append(node.Children, parseStatement(p)...)

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

func parseIf(p *parser) *Node {
	node := &Node{
		Kind:      KindIf,
		StartLine: p.peek().StartLine,
		EndLine:   p.peek().EndLine,
	}

	p.expect(lexer.KindIf)
	p.expect(lexer.KindSpace)
	p.skipWhitespace()

	// TODO validate this returns a KindInfix, or KindNot
	node.Children = append(node.Children, parseExpression(p))
	p.skipWhitespace()

	// happy path (if case)
	node.Children = append(node.Children, parseMany(p)...)

	p.skipWhitespace()

	if p.peek().Kind == lexer.KindElse {
		p.expect(lexer.KindElse)
		p.skipWhitespace()
		p.expect(lexer.KindRightDelim)
		// sad path (else case)
		node.Children = append(node.Children, parseMany(p)...)
		p.skipWhitespace()
	}

	p.expect(lexer.KindEnd)

	return node
}

func parseOperator(p *parser) *Node {
	token := p.next()
	node := &Node{
		Kind:      KindOperator,
		Value:     token.Value,
		StartLine: token.StartLine,
	}

	token = p.expect(lexer.KindEqual)
	node.Value += "="
	node.EndLine = token.EndLine

	return node
}
