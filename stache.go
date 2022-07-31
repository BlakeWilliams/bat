package stache

import (
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/blakewilliams/stache/internal/lexer"
	"github.com/blakewilliams/stache/internal/parser"
)

type Template struct {
	Name string
	ast  *parser.Node
}

func New(input string) Template {
	l := lexer.Lex(input)
	ast := parser.Parse(l)

	return Template{ast: ast}
}

func (t *Template) Execute(out io.Writer, data map[string]any) (err error) {
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

	for _, child := range t.ast.Children {
		eval(child, out, data)
	}

	return nil
}

func eval(n *parser.Node, out io.Writer, data map[string]any) {
	switch n.Kind {
	case parser.KindText:
		out.Write([]byte(n.Value))
	case parser.KindStatement:
		eval(n.Children[0], out, data)
	case parser.KindAccess:
		value := access(n, data)

		out.Write([]byte(valueToString(value)))
	case parser.KindIdentifier:
		value := access(n, data)

		out.Write([]byte(valueToString(value)))
	default:
		panic(fmt.Sprintf("unsupported kind %s", n.Kind))
	}
}

func access(n *parser.Node, data map[string]any) any {
	switch n.Kind {
	case parser.KindIdentifier:
		return data[n.Value]
	case parser.KindAccess:
		root := access(n.Children[0], data)
		propName := n.Children[1].Value

		v := reflect.ValueOf(root)

		k := v.Kind()
		if k == reflect.Pointer {
			v = v.Elem()
			k = v.Kind()
		}

		switch k {
		case reflect.Struct:
			value := reflect.Indirect(v).FieldByName(propName)
			return value.Interface()
		case reflect.Map:
			value := v.MapIndex(reflect.ValueOf(propName))
			return value.Interface()
		default:
			panic(fmt.Sprintf("unsupported dot access on type %s", k))
			// panic(fmt.Sprintf("unsupported dot access on type %s on line %d column %d", k, n.Line, n.Column))
		}
	default:
		panic(fmt.Sprintf("unsupported access called on type %s", n.Kind))
	}
}

func valueToString(v any) string {
	return fmt.Sprintf("%v", v)
}
