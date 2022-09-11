package bat

import (
	"errors"
	"fmt"
	"html"
	"io"
	"reflect"
	"strconv"
	"strings"

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
	raw        string
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

	t := Template{raw: input, ast: ast, escapeFunc: HTMLEscape}
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
		t.eval(child, out, data, t.helpers, make(map[string]any))
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

func (t *Template) eval(n *parser.Node, out io.Writer, data map[string]any, helpers map[string]any, vars map[string]any) {
	switch n.Kind {
	case parser.KindText:
		out.Write([]byte(n.Value))
	case parser.KindNot:
		value := t.access(n, data, helpers, vars)
		out.Write([]byte(valueToString(value, t.escapeFunc)))
	case parser.KindString:
		out.Write([]byte(n.Value)[1 : len(n.Value)-1])
	case parser.KindStatement:
		t.eval(n.Children[0], out, data, helpers, vars)
	case parser.KindAccess, parser.KindNegate, parser.KindBracketAccess:
		value := t.access(n, data, helpers, vars)

		out.Write([]byte(valueToString(value, t.escapeFunc)))
	case parser.KindIdentifier, parser.KindVariable, parser.KindInt, parser.KindInfix, parser.KindCall, parser.KindMap:
		value := t.access(n, data, helpers, vars)

		out.Write([]byte(valueToString(value, t.escapeFunc)))
	case parser.KindIf:
		conditionResult := t.access(n.Children[0], data, helpers, vars)
		v := reflect.ValueOf(conditionResult)

		if isTruthy(v) {
			t.eval(n.Children[1], out, data, helpers, vars)
		} else if len(n.Children) > 2 && n.Children[2] != nil {
			t.eval(n.Children[2], out, data, helpers, vars)
		}
	case parser.KindBlock:
		for _, child := range n.Children {
			t.eval(child, out, data, helpers, vars)
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
			toLoop = t.access(n.Children[2], data, helpers, vars)
			body = n.Children[3]
		} else {
			toLoop = t.access(n.Children[1], data, helpers, vars)
			body = n.Children[2]
		}

		v := reflect.ValueOf(toLoop)

		switch v.Kind() {
		case reflect.Slice, reflect.Array:
			for i := 0; i < v.Len(); i++ {
				newVars[iteratorName] = i
				newVars[valueName] = v.Index(i).Interface()

				t.eval(body, out, data, helpers, newVars)
			}
		case reflect.Map:
			sorted := mapsort.Sort(v)

			for i := range sorted.Keys {
				newVars[iteratorName] = sorted.Keys[i].Interface()
				newVars[valueName] = sorted.Values[i].Interface()

				t.eval(body, out, data, helpers, newVars)
			}
		case reflect.Chan:
			defaultCase := reflect.SelectCase{Dir: reflect.SelectDefault}
			recvCase := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: v}

			i := 0
			cases := []reflect.SelectCase{defaultCase, recvCase}
			for {
				chosen, value, ok := reflect.Select(cases)

				if chosen == 0 || !ok {
					break
				}
				newVars[iteratorName] = i
				newVars[valueName] = value.Interface()
				t.eval(body, out, data, helpers, newVars)
				i++
			}
		default:
			t.panicWithTrace(n, fmt.Sprintf("attempted to range over %s", v.Kind()))
		}
	default:
		t.panicWithTrace(n, fmt.Sprintf("unsupported kind %s", n.Kind))
	}
}

func (t *Template) access(n *parser.Node, data map[string]any, helpers map[string]any, vars map[string]any) any {
	switch n.Kind {
	case parser.KindCall:
		toCall := reflect.ValueOf(t.access(n.Children[0], data, helpers, vars))
		args := make([]reflect.Value, 0, len(n.Children)-1)
		for _, arg := range n.Children[1:] {
			args = append(args, reflect.ValueOf(t.access(arg, data, helpers, vars)))
		}

		// Wrap the call in a closure to allow for the possibility of panics so
		// we can provide good error messages
		return func() any {
			defer func() {
				if err := recover(); err != nil {
					t.panicWithTrace(n.Children[0], fmt.Sprintf("error calling function '%s': %s", n.Children[0].Value, err))
				}
			}()

			return toCall.Call(args)[0].Interface()
		}()
	case parser.KindNegate:
		value := t.access(n.Children[0], data, helpers, vars)
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
			t.panicWithTrace(n, fmt.Sprintf("can't negate type %s", reflect.ValueOf(value).Kind()))
			return nil
		}
	case parser.KindNot:
		value := t.access(n.Children[0], data, helpers, vars)

		if value == nil || value == false {
			return true
		} else {
			return false
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
		left := t.access(n.Children[0], data, helpers, vars)
		right := t.access(n.Children[2], data, helpers, vars)

		switch n.Children[1].Value {
		case "!=":
			return !compare(reflect.ValueOf(left), reflect.ValueOf(right))
		case "==":
			return compare(reflect.ValueOf(left), reflect.ValueOf(right))
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
		case "<":
			return lessThan(left, right)
		case ">":
			return greaterThan(left, right)
		case "<=":
			return lessThan(left, right) || compare(reflect.ValueOf(left), reflect.ValueOf(right))
		case ">=":
			return greaterThan(left, right) || compare(reflect.ValueOf(left), reflect.ValueOf(right))
		default:
			t.panicWithTrace(n, fmt.Sprintf("Unsupported operator '%s'", n.Children[1].Value))
			return nil
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
	case parser.KindMap:
		m := make(map[string]any, len(n.Children))

		for _, child := range n.Children {
			key := child.Children[0]
			value := child.Children[1]

			m[key.Value] = reflect.ValueOf(t.access(value, data, helpers, vars)).Interface()
		}

		return m
	case parser.KindBracketAccess:
		root := t.access(n.Children[0], data, helpers, vars)
		accessor := t.access(n.Children[1], data, helpers, vars)

		rootVal := reflect.ValueOf(root)
		accessorVal := reflect.ValueOf(accessor)

		switch rootVal.Kind() {
		case reflect.Map:
			return rootVal.MapIndex(reflect.ValueOf(accessor)).Interface()
		case reflect.Slice, reflect.Array:
			switch accessorVal.Kind() {
			case reflect.Int:
				return rootVal.Index(accessor.(int)).Interface()
			case reflect.Int16:
				return rootVal.Index(int(accessor.(int16))).Interface()
			case reflect.Int32:
				return rootVal.Index(int(accessor.(int32))).Interface()
			case reflect.Int64:
				return rootVal.Index(int(accessor.(int64))).Interface()
			case reflect.Uint:
				return rootVal.Index(int(accessor.(uint))).Interface()
			case reflect.Uint16:
				return rootVal.Index(int(accessor.(uint16))).Interface()
			case reflect.Uint32:
				return rootVal.Index(int(accessor.(uint32))).Interface()
			case reflect.Uint64:
				return rootVal.Index(int(accessor.(uint64))).Interface()
			default:
				t.panicWithTrace(n, fmt.Sprintf("can't index %s with %s", rootVal.Kind(), accessorVal.Kind()))
				return nil
			}
		default:
			t.panicWithTrace(n, "cannot index non-map/non-slice")
			return nil
		}
	case parser.KindAccess:
		root := t.access(n.Children[0], data, helpers, vars)
		propName := n.Children[1].Value

		if root == nil {
			t.panicWithTrace(n, fmt.Sprintf("attempted to access property `%s` on nil value on line %d", propName, n.StartLine))
			return nil
		}

		v := reflect.ValueOf(root)
		k := v.Kind()

		// Special case structs, because pointer methods
		if k == reflect.Struct || k == reflect.Pointer && v.Elem().Kind() == reflect.Struct {
			// Support field access
			if value := reflect.Indirect(v).FieldByName(propName); !reflect.ValueOf(value).IsZero() {
				return value.Interface()
			}

			// Support method access
			if value := v.MethodByName(propName); !reflect.ValueOf(value).IsZero() {
				return value.Interface()
			}

			t.panicWithTrace(n, fmt.Sprintf("no field or method '%s' for type %s on line %d", propName, reflect.TypeOf(root), n.StartLine))
			return nil
		}

		if k == reflect.Pointer {
			v = v.Elem()
			k = v.Kind()
		}

		switch k {
		case reflect.Map:
			value := v.MapIndex(reflect.ValueOf(propName))
			return value.Interface()
		default:
			t.panicWithTrace(n, fmt.Sprintf("access on type %s on line %d", k, n.StartLine))
			return nil
		}
	case parser.KindString:
		// Cut off opening " and closing "
		return n.Value[1 : len(n.Value)-1]
	default:
		t.panicWithTrace(n, fmt.Sprintf("unsupported access called on type %s", n.Kind))
		return nil
	}
}

func (t *Template) panicWithTrace(n *parser.Node, msg string) {
	lines := strings.Split(t.raw, "\n")

	endLine := n.EndLine
	if endLine == 0 {
		endLine = n.StartLine
	}
	relevantLines := lines[n.StartLine-1 : endLine]

	errorMessage := fmt.Sprintf("%s starting on line %d:\n%s", msg, n.StartLine, strings.Join(relevantLines, "\n"))

	panic(errorMessage)
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
	case nil:
		return ""
	default:
		return escape(fmt.Sprintf("%v", v))
	}
}
