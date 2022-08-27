package parser

import (
	"errors"
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
	KindTrue       = "true"
	KindFalse      = "false"
	KindRange      = "range"
	KindVariable   = "variable"
	KindString     = "string"
	KindInt        = "int"
	// Represents blocks of top-level items for use in if/else, range body, etc.
	KindBlock         = "block"
	KindNegate        = "negate"
	KindCall          = "call"
	KindMap           = "map"
	KindPair          = "pair"
	KindBracketAccess = "bracket_access"
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

func Parse(l *lexer.Lexer) (_ *Node, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch val := r.(type) {
			case string:
				err = errors.New(val)
			case error:
				err = val
			}
		}
	}()

	p := &parser{
		lexer: l,
		Root:  &Node{Kind: KindRoot},
		pos:   -1,
	}

	p.Root.Children = parseMany(p)

	return p.Root, err
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
			node.Children = []*Node{parseStatement(p)}
			nodes = append(nodes, node)
			p.skipWhitespace()
			p.expect(lexer.KindRightDelim)
		case lexer.KindElse:
			return nodes
		case lexer.KindEnd:
			return nodes
		default:
			p.errorWithLoc("unsupported token %v", p.peek().Value)
		}
	}
}

// Statements represent everything in a `{{...}}` block.
func parseStatement(p *parser) *Node {
	p.skipWhitespace()

	switch p.peek().Kind {
	case lexer.KindRightDelim:
		p.next()
	case lexer.KindEOF:
		panic("unexpected EOF")
	case lexer.KindOpenCurly, lexer.KindIdentifier, lexer.KindVariable, lexer.KindNumber, lexer.KindMinus:
		return parseExpression(p, true)
	case lexer.KindNil:
		token := p.next()
		return &Node{Kind: KindNil, StartLine: token.StartLine, EndLine: token.EndLine}
	case lexer.KindSpace:
		p.skipWhitespace()
		return nil
	case lexer.KindIf:
		return parseIf(p)
	case lexer.KindRange:
		return parseRange(p)
	default:
		p.errorWithLoc("unexpected token %v", p.peek().Value)
	}
	return nil
}

func (p *parser) errorWithLoc(msg string, formatting ...any) {
	formatted := fmt.Sprintf(msg, formatting...)
	formatted += fmt.Sprintf(": on line %d", p.peek().StartLine)

	panic(formatted)
}

// parses expressions, like:
// foo.bar.baz
// foo != nil
func parseExpression(p *parser, allowOperator bool) *Node {
	var rootNode *Node
	if p.peek().Kind == lexer.KindOpenCurly {
		p.expect(lexer.KindOpenCurly)
		rootNode = parseMap(p)
	} else {
		rootNode = parseLiteralOrAccess(p)
	}

	p.skipWhitespace()

	if p.peek().Kind == lexer.KindDot || p.peek().Kind == lexer.KindOpenParen || p.peek().Kind == lexer.KindOpenBracket {
		node := rootNode

	loop:
		for {
			switch p.peek().Kind {
			case lexer.KindDot:
				p.expect(lexer.KindDot)
				childNode := parseVariable(p)

				newNode := &Node{
					Kind:      KindAccess,
					Children:  []*Node{node, childNode},
					StartLine: childNode.StartLine,
					EndLine:   childNode.EndLine,
				}

				node = newNode
			case lexer.KindOpenBracket:
				p.expect(lexer.KindOpenBracket)

				newNode := &Node{
					Kind:      KindBracketAccess,
					Children:  []*Node{node},
					StartLine: rootNode.StartLine,
				}

				child := parseExpression(p, true)
				newNode.Children = append(newNode.Children, child)
				p.expect(lexer.KindCloseBracket)

				node = newNode
			case lexer.KindOpenParen:
				p.expect(lexer.KindOpenParen)
				newNode := &Node{
					Kind:      KindCall,
					Children:  []*Node{node},
					StartLine: rootNode.StartLine,
				}

				for {
					p.skipWhitespace()
					if p.peek().Kind == lexer.KindCloseParen {
						break
					}

					newNode.Children = append(newNode.Children, parseExpression(p, true))

					if p.peek().Kind == lexer.KindComma {
						p.expect(lexer.KindComma)
					}
				}

				p.expect(lexer.KindCloseParen)

				node = newNode
			default:
				break loop
			}
		}

		return node
	}

	if !allowOperator {
		return rootNode
	}

	// check for ==, -, !=,
	// protect against foo -1 vs foo - 1 and foo != bar vs foo !bar
	next := p.peek()
	switch next.Kind {
	case lexer.KindMinus:
		if p.peekn(2).Kind != lexer.KindSpace {
			return rootNode
		}
	case lexer.KindBang:
		if p.peekn(2).Kind != lexer.KindEqual {
			return rootNode
		}
	case lexer.KindEqual, lexer.KindPlus, lexer.KindSlash, lexer.KindAsterisk, lexer.KindPercent:
		// do nothing, fall through to parse operator
	default:
		return rootNode
	}

	operator := parseOperator(p)
	p.skipWhitespace()

	node := &Node{
		Kind:      KindInfix,
		Children:  []*Node{},
		StartLine: rootNode.StartLine,
		EndLine:   p.peek().EndLine,
	}

	node.Children = append(node.Children, rootNode)
	node.Children = append(node.Children, operator)
	right := parseExpression(p, false)

	if right.Kind == KindInfix {
		panic("infix operator cannot follow infix operator")
	}
	node.Children = append(node.Children, right)

	return node
}

func parseLiteralOrAccess(p *parser) *Node {
	kind := KindIdentifier
	switch p.peek().Kind {
	case lexer.KindNil:
		kind = KindNil
	case lexer.KindTrue:
		kind = KindTrue
	case lexer.KindFalse:
		kind = KindFalse
	case lexer.KindString:
		kind = KindString
	case lexer.KindMinus:
		switch p.peekn(2).Kind {
		case lexer.KindNumber:
			kind = KindInt

			p.next()
			intNode := p.next()
			p.skipWhitespace() // copy whitespace skipping logic below before return

			return &Node{
				Kind:      kind,
				Value:     "-" + intNode.Value,
				StartLine: intNode.StartLine,
				EndLine:   intNode.EndLine,
			}
		case lexer.KindVariable, lexer.KindIdentifier:
			p.next()
			p.skipWhitespace()
			return &Node{
				Kind:      KindNegate,
				StartLine: p.peek().StartLine,
				EndLine:   p.peek().EndLine,
				Children:  []*Node{parseExpression(p, true)},
			}
		default:
			panic(fmt.Sprintf("Unexpected token `-` on line %d", p.peek().StartLine))
		}
	case lexer.KindNumber:
		kind = KindInt
	case lexer.KindVariable, lexer.KindIdentifier:
		return parseVariable(p)
	default:
		panic(fmt.Sprintf("Unexpected identifier %s", p.peek().Kind.String()))
	}

	identifierToken := p.next()

	identifierNode := &Node{
		Kind:      kind,
		Value:     identifierToken.Value,
		StartLine: identifierToken.StartLine,
		EndLine:   identifierToken.EndLine,
	}

	p.skipWhitespace()

	return identifierNode
}

func parseVariable(p *parser) *Node {
	identifierToken := p.next()

	kind := KindIdentifier
	switch identifierToken.Kind {
	case lexer.KindVariable:
		kind = KindVariable
	case lexer.KindIdentifier:
		kind = KindIdentifier
	default:
		panic(fmt.Sprintf("unexpected token %s, expected variable or identifier", identifierToken.Value))
	}

	rootNode := &Node{
		Kind:      kind,
		Value:     identifierToken.Value,
		StartLine: identifierToken.StartLine,
		EndLine:   identifierToken.EndLine,
	}

	return rootNode
}

func (p *parser) expect(kind lexer.Kind) lexer.Token {
	n := p.next()

	if n.Kind != kind {
		panic(fmt.Sprintf("unexpected token %v, expected %s", n.Value, kind))
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
	node.Children = append(node.Children, parseExpression(p, true))
	p.skipWhitespace()
	p.expect(lexer.KindRightDelim)

	// happy path (if case)
	node.Children = append(node.Children, parseBlock(p))

	p.skipWhitespace()

	if p.peek().Kind == lexer.KindElse {
		p.expect(lexer.KindElse)
		p.skipWhitespace()
		p.expect(lexer.KindRightDelim)
		// sad path (else case)
		node.Children = append(node.Children, parseBlock(p))
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

	if token.Kind == lexer.KindEqual || token.Kind == lexer.KindBang {
		token = p.expect(lexer.KindEqual)
		node.Value += "="
	} else {
		p.skipWhitespace()
	}
	node.EndLine = token.EndLine

	return node
}

func parseRange(p *parser) *Node {
	rangeToken := p.expect(lexer.KindRange)
	node := &Node{
		Kind:      KindRange,
		StartLine: rangeToken.StartLine,
		Children:  make([]*Node, 0, 3),
	}

	p.skipWhitespace()
	var1Token := p.expect(lexer.KindVariable)
	var1 := &Node{
		Kind:      KindVariable,
		StartLine: rangeToken.StartLine,
		EndLine:   rangeToken.EndLine,
		Value:     var1Token.Value,
	}
	node.Children = append(node.Children, var1)
	p.skipWhitespace()

	if p.peek().Kind == lexer.KindComma {
		p.next()
		p.skipWhitespace()
		var2Token := p.expect(lexer.KindVariable)
		var2 := &Node{
			Kind:      KindVariable,
			StartLine: var2Token.StartLine,
			EndLine:   var2Token.EndLine,
			Value:     var2Token.Value,
		}
		node.Children = append(node.Children, var2)
	}
	p.skipWhitespace()
	p.expect(lexer.KindIn)
	p.skipWhitespace()

	node.Children = append(node.Children, parseExpression(p, true))
	p.expect(lexer.KindRightDelim)
	node.Children = append(node.Children, parseBlock(p))
	p.skipWhitespace()
	p.expect(lexer.KindEnd)

	return node
}

func parseBlock(p *parser) *Node {
	startToken := p.peek()
	node := &Node{
		Kind:      KindBlock,
		StartLine: startToken.StartLine,
		EndLine:   startToken.EndLine, // TODO fix
		Children:  make([]*Node, 0),
	}

	node.Children = append(node.Children, parseMany(p)...)

	return node
}

func parseMap(p *parser) *Node {
	p.skipWhitespace()
	mapNode := &Node{
		Kind:      KindMap,
		StartLine: p.peek().StartLine,
	}

	pairs := make([]*Node, 0)
	for {
		if p.peek().Kind == lexer.KindCloseCurly {
			break
		}

		if p.peek().Kind == lexer.KindEOF {
			p.errorWithLoc("unexpected EOF")
		}

		key := p.expect(lexer.KindIdentifier)
		p.expect(lexer.KindColon)
		p.skipWhitespace()
		value := parseLiteralOrAccess(p)

		pair := &Node{
			Kind: KindPair,
			Children: []*Node{
				{Kind: KindIdentifier, Value: key.Value, StartLine: key.StartLine, EndLine: key.EndLine},
				value,
			},
			StartLine: key.StartLine,
			EndLine:   value.EndLine,
		}

		pairs = append(pairs, pair)

		// check for comma
		p.skipWhitespace()
		if p.peek().Kind == lexer.KindComma {
			p.expect(lexer.KindComma)
			p.skipWhitespace()
		}
	}

	mapNode.Children = pairs
	mapNode.EndLine = pairs[len(pairs)-1].EndLine

	p.skipWhitespace()
	p.expect(lexer.KindCloseCurly)

	return mapNode
}
