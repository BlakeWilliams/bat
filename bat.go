package bat

import (
	"errors"
	"fmt"
	"html"
	"io"
	"reflect"
	"strconv"

	"github.com/blakewilliams/bat/internal/lexer"
	"github.com/blakewilliams/bat/internal/mapsort"
	"github.com/blakewilliams/bat/internal/parser"
)

// Represents a single template that can be rendered.
type Template struct {
	Name       string
	ast        *parser.Node
	helpers    map[string]any
	escapeFunc func(string) string
}

// An escapeFunc that returns text as-is
func NoEscape(s string) string { return s }

// An escapeFunc that returns text as escaped HTML
var HTMLEscape func(s string) string = html.EscapeString

// Safe values are not escaped. These should be used carefully as they expose
// risk to your templates outputting unsafe values, especially if the values
// are derived from user input.
type Safe string

// A function that allows the template to be customized when using NewTemplate.
type TemplateOption = func(*Template)

// Creates a new template using the provided input. Options can be provided to
// customize the template, such as setting the function used to escape unsafe
// input.
func NewTemplate(input string, opts ...TemplateOption) (Template, error) {
	l := lexer.Lex(input)
	ast, err := parser.Parse(l)

	if err != nil {
		return Template{}, fmt.Errorf("could not create template: %w", err)
	}

	t := Template{ast: ast, escapeFunc: HTMLEscape}
	for _, opt := range opts {
		opt(&t)
	}

	return t, nil
}

// Executes the template, streaming output to out. The data parameter is made
// available to the template.
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
		eval(child, t.escapeFunc, out, data, t.helpers, make(map[string]any))
	}

	return nil
}

// An option function that provides a custom escape function that is used to
// escape unsafe dynamic template values.
func WithEscapeFunc(fn func(string) string) func(*Template) {
	return func(t *Template) {
		t.escapeFunc = fn
	}
}

func WithHelpers(fns map[string]any) TemplateOption {
	return func(t *Template) {
		t.helpers = fns
	}
}

func eval(n *parser.Node, escapeFunc func(string) string, out io.Writer, data map[string]any, helpers map[string]any, vars map[string]any) {
	switch n.Kind {
	case parser.KindText:
		out.Write([]byte(n.Value))
	case parser.KindStatement:
		eval(n.Children[0], escapeFunc, out, data, helpers, vars)
	case parser.KindAccess, parser.KindNegate:
		value := access(n, data, helpers, vars)

		out.Write([]byte(valueToString(value, escapeFunc)))
	case parser.KindIdentifier, parser.KindVariable, parser.KindInt, parser.KindInfix, parser.KindCall:
		value := access(n, data, helpers, vars)

		out.Write([]byte(valueToString(value, escapeFunc)))
	case parser.KindIf:
		conditionResult := access(n.Children[0], data, helpers, vars)

		if conditionResult == true {
			eval(n.Children[1], escapeFunc, out, data, helpers, vars)
		} else if n.Children[2] != nil {
			eval(n.Children[2], escapeFunc, out, data, helpers, vars)
		}
	case parser.KindBlock:
		for _, child := range n.Children {
			eval(child, escapeFunc, out, data, helpers, vars)
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
			toLoop = access(n.Children[2], data, helpers, vars)
			body = n.Children[3]
		} else {
			toLoop = access(n.Children[1], data, helpers, vars)
			body = n.Children[2]
		}

		v := reflect.ValueOf(toLoop)

		switch v.Kind() {
		case reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				newVars[iteratorName] = i
				newVars[valueName] = v.Index(i).Interface()

				eval(body, escapeFunc, out, data, helpers, newVars)
			}
		case reflect.Map:
			sorted := mapsort.Sort(v)

			for i := range sorted.Keys {
				newVars[iteratorName] = sorted.Keys[i].Interface()
				newVars[valueName] = sorted.Values[i].Interface()

				eval(body, escapeFunc, out, data, helpers, newVars)
			}
		default:
			panic(fmt.Sprintf("attempted to range over %s", v.Kind()))
		}
	default:
		panic(fmt.Sprintf("unsupported kind %s", n.Kind))
	}
}

func access(n *parser.Node, data map[string]any, helpers map[string]any, vars map[string]any) any {
	switch n.Kind {
	case parser.KindCall:
		toCall := reflect.ValueOf(access(n.Children[0], data, helpers, vars))
		args := make([]reflect.Value, 0, len(n.Children)-1)
		for _, arg := range n.Children[1:] {
			args = append(args, reflect.ValueOf(access(arg, data, helpers, vars)))
		}

		return toCall.Call(args)[0].Interface()
	case parser.KindNegate:
		value := access(n.Children[0], data, helpers, vars)
		switch reflect.ValueOf(value).Kind() {
		case reflect.Int:
			return value.(int) * -1
		case reflect.Int16:
			return value.(int16) * -1
		case reflect.Int32:
			return value.(int32) * -1
		case reflect.Int64:
			return value.(int64) * -1
		case reflect.Complex64:
			return value.(complex64) * -1
		case reflect.Complex128:
			return value.(complex128) * -1
		case reflect.Float32:
			return value.(float32) * -1
		case reflect.Float64:
			return value.(float64) * -1
		default:
			panic(fmt.Sprintf("can't negate type %s", reflect.ValueOf(value).Kind()))
		}
	case parser.KindTrue:
		return true
	case parser.KindFalse:
		return false
	case parser.KindNil:
		return nil
	case parser.KindInt:
		val, _ := strconv.Atoi(n.Value)
		return val
	case parser.KindInfix:
		left := access(n.Children[0], data, helpers, vars)
		right := access(n.Children[2], data, helpers, vars)

		switch n.Children[1].Value {
		case "!=":
			return left != right
		case "==":
			return left == right
		case "-":
			return subtract(left, right)
		case "+":
			return add(left, right)
		case "*":
			return multiply(left, right)
		case "/":
			return divide(left, right)
		case "%":
			return modulo(left, right)
		default:
			panic(fmt.Sprintf("Unsupported operator: %s on line %d", n.Children[1].Value, n.Children[1].StartLine))
		}

	case parser.KindIdentifier:
		if val, ok := data[n.Value]; ok {
			return val
		}

		if val, ok := helpers[n.Value]; ok {
			return val
		}

		return nil
	case parser.KindVariable:
		return vars[n.Value]
	case parser.KindAccess:
		root := access(n.Children[0], data, helpers, vars)
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
			// Support field access
			if value := reflect.Indirect(v).FieldByName(propName); !reflect.ValueOf(value).IsZero() {
				return value.Interface()
			}

			// Support method access
			if value := reflect.Indirect(v).MethodByName(propName); !reflect.ValueOf(value).IsZero() {
				return value.Interface()
			}

			return nil
		case reflect.Map:
			value := v.MapIndex(reflect.ValueOf(propName))
			return value.Interface()
		default:
			panic(fmt.Sprintf("access on type %s on line %d", k, n.StartLine))
		}
	case parser.KindString:
		// Cut off opening " and closing "
		return n.Value[1 : len(n.Value)-1]
	default:
		panic(fmt.Sprintf("unsupported access called on type %s", n.Kind))
	}
}

// TODO this needs to check for the stringer interface, and maybe handle values
// a bit more gracefully...
func valueToString(v any, escape func(string) string) string {
	if val, ok := v.(fmt.Stringer); ok {
		return escape(val.String())
	}

	switch val := v.(type) {
	case Safe:
		return string(val)
	case string:
		return escape(val)
	default:
		return escape(fmt.Sprintf("%v", v))
	}
}
