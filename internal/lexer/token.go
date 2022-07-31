package lexer

import "fmt"

const (
	KindError Kind = iota
	// Represents raw text in a template
	KindText
	KindEOF
	KindLeftDelim
	KindRightDelim
	KindIdentifier
	KindDot
	KindHash
	KindSpace
	KindBang
	KindEqual
	KindIf
	KindNil
	KindElse
	KindEnd
)

type Token struct {
	Kind      Kind
	Value     string
	StartLine int
	EndLine   int
}

func (k Kind) String() string {
	switch k {
	case KindError:
		return "error"
	case KindText:
		return "text"
	case KindEOF:
		return "eof"
	case KindLeftDelim:
		return "openDelim"
	case KindRightDelim:
		return "closeDelim"
	case KindIdentifier:
		return "ident"
	case KindDot:
		return "dot"
	case KindHash:
		return "hash"
	case KindSpace:
		return "space"
	case KindBang:
		return "bang"
	case KindEqual:
		return "equal"
	case KindIf:
		return "if"
	case KindNil:
		return "nil"
	case KindElse:
		return "else"
	case KindEnd:
		return "end"
	default:
		return fmt.Sprintf("uknown %d", k)
	}
}

func (t Token) String() string {
	return fmt.Sprintf("{%s `%s`}", t.Kind, t.Value)
}
