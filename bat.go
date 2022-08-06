package bat

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"

	"github.com/blakewilliams/bat/internal/lexer"
	"github.com/blakewilliams/bat/internal/mapsort"
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
	case parser.KindAccess, parser.KindNegate:
		value := access(n, data, vars)

		out.Write([]byte(valueToString(value)))
	case parser.KindIdentifier, parser.KindVariable, parser.KindInt, parser.KindInfix:
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
				newVars[valueName] = v.Index(i).Interface()

				eval(body, out, data, newVars)
			}
		case reflect.Map:
			sorted := mapsort.Sort(v)

			for i := range sorted.Keys {
				newVars[iteratorName] = sorted.Keys[i].Interface()
				newVars[valueName] = sorted.Values[i].Interface()

				eval(body, out, data, newVars)
			}
		default:
			panic(fmt.Sprintf("attempted to range over %s", v.Kind()))
		}
	default:
		panic(fmt.Sprintf("unsupported kind %s", n.Kind))
	}
}

func access(n *parser.Node, data map[string]any, vars map[string]any) any {
	switch n.Kind {
	case parser.KindNegate:
		value := access(n.Children[0], data, vars)
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
		left := access(n.Children[0], data, vars)
		right := access(n.Children[2], data, vars)

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
	case parser.KindString:
		// Cut off opening " and closing "
		return n.Value[1 : len(n.Value)-1]
	default:
		panic(fmt.Sprintf("unsupported access called on type %s", n.Kind))
	}
}

// TODO this needs to check for the stringer interface, and maybe handle values
// a bit more gracefully...
func valueToString(v any) string {
	return fmt.Sprintf("%v", v)
}
