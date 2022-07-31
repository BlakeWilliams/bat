package bat

import (
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/blakewilliams/bat/internal/lexer"
	"github.com/blakewilliams/bat/internal/parser"
)

type Template struct {
	Name string
	ast  *parser.Node
}

func NewTemplate(input string) (Template, error) {
	l := lexer.Lex(input)
	ast, err := parser.Parse(l)

	if err != nil {
		return Template{}, fmt.Errorf("could not create template: %w", err)
	}

	return Template{ast: ast}, nil
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
		eval(child, out, data, make(map[string]any))
	}

	return nil
}

func eval(n *parser.Node, out io.Writer, data map[string]any, vars map[string]any) {
	switch n.Kind {
	case parser.KindText:
		out.Write([]byte(n.Value))
	case parser.KindStatement:
		eval(n.Children[0], out, data, vars)
	case parser.KindAccess:
		value := access(n, data, vars)

		out.Write([]byte(valueToString(value)))
	case parser.KindIdentifier, parser.KindVariable:
		value := access(n, data, vars)

		out.Write([]byte(valueToString(value)))
	case parser.KindIf:
		conditionResult := access(n.Children[0], data, vars)

		if conditionResult == true {
			eval(n.Children[1], out, data, vars)
		} else if n.Children[2] != nil {
			eval(n.Children[2], out, data, vars)
		}
	case parser.KindBlock:
		for _, child := range n.Children {
			eval(child, out, data, vars)
		}
	case parser.KindRange:
		newVars := make(map[string]any, len(vars)+2)
		for k, v := range vars {
			newVars[k] = v
		}

		iteratorName := n.Children[0].Value
		valueName := n.Children[1].Value

		var toLoop any
		var body *parser.Node

		if len(n.Children) == 4 {
			toLoop = access(n.Children[2], data, vars)
			body = n.Children[3]
		} else {
			toLoop = access(n.Children[1], data, vars)
			body = n.Children[2]
		}

		v := reflect.ValueOf(toLoop)

		switch v.Kind() {
		case reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				newVars[iteratorName] = i
				newVars[valueName] = v.Index(i)

				eval(body, out, data, newVars)
			}
		case reflect.Map:
			for _, key := range v.MapKeys() {
				newVars[iteratorName] = key
				newVars[valueName] = v.MapIndex(key)

				eval(body, out, data, newVars)
			}
		default:
			panic(fmt.Sprintf("attempted to range over %s", v.Kind()))
		}
		// TODO validate we can loop
	default:
		panic(fmt.Sprintf("unsupported kind %s", n.Kind))
	}
}

func access(n *parser.Node, data map[string]any, vars map[string]any) any {
	switch n.Kind {
	case parser.KindTrue:
		return true
	case parser.KindFalse:
		return false
	case parser.KindNil:
		return nil
	case parser.KindInfix:
		left := access(n.Children[0], data, vars)
		right := access(n.Children[2], data, vars)

		if n.Children[1].Value == "!=" {
			return left != right
		} else if n.Children[1].Value == "==" {
			return left == right
		} else {
			panic(fmt.Sprintf("Unsupported operator: %s on line %d", n.Children[1].Value, n.Children[1].StartLine))
		}
	case parser.KindIdentifier:
		return data[n.Value]
	case parser.KindVariable:
		return vars[n.Value]
	case parser.KindAccess:
		root := access(n.Children[0], data, vars)
		propName := n.Children[1].Value

		if root == nil {
			panic(fmt.Sprintf("attempted to access property `%s` on nil value on line %d", propName, n.StartLine))
		}

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
			panic(fmt.Sprintf("access on type %s on line %d", k, n.StartLine))
		}
	default:
		panic(fmt.Sprintf("unsupported access called on type %s", n.Kind))
	}
}

// TODO this needs to check for the stringer interface, and maybe handle values
// a bit more gracefully...
func valueToString(v any) string {
	return fmt.Sprintf("%v", v)
}
