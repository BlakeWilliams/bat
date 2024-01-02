package bat

import (
	"bytes"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemplate(t *testing.T) {
	template, err := NewTemplate("hello.html", "<h1>Hello {{name}}</h1>")
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, nil, map[string]any{"name": "Fox Mulder"})
	require.NoError(t, err)

	require.Equal(t, "<h1>Hello Fox Mulder</h1>", b.String())
}

type user struct {
	Name name
}

func (u user) GetName() name {
	return u.Name
}

type name struct {
	First string
	Last  string
}

func (n name) Initials() string {
	return n.First[0:1] + n.Last[0:1]
}

func TestTemplateDots(t *testing.T) {
	user := user{
		Name: name{First: "Fox", Last: "Mulder"},
	}

	template, err := NewTemplate("hello.html", "<h1>Hello {{user.Name.First}} {{user.Name.Last}}</h1>")
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, nil, map[string]any{"user": user})
	require.NoError(t, err)

	require.Equal(t, "<h1>Hello Fox Mulder</h1>", b.String())

}

func TestTemplateDots_Map(t *testing.T) {
	user := map[string]map[string]string{
		"user": {
			"name": "Fox Mulder",
		},
	}

	template, err := NewTemplate("hello.html", "<h1>Hello {{details.user.name}}</h1>")
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, nil, map[string]any{"details": user})
	require.NoError(t, err)

	require.Equal(t, "<h1>Hello Fox Mulder</h1>", b.String())
}

func TestTemplateDotsNil(t *testing.T) {
	template, err := NewTemplate("hello.html", "<h1>Hello {{details.user.name}}</h1>")
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, nil, map[string]any{})
	require.Error(t, err)
	require.ErrorContains(t, err, "attempted to access property `user` on nil value")
	require.ErrorContains(t, err, "on line 1")
}

func TestTemplate_If_Spacing(t *testing.T) {
	template, err := NewTemplate("hello.html", "{{ if name != nil }}Hello!{{ else }}Goodbye!{{ end }}")
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, nil, map[string]any{"name": "Fox Mulder"})
	require.NoError(t, err)

	require.Equal(t, "Hello!", b.String())

	b = new(bytes.Buffer)
	err = template.Execute(b, nil, map[string]any{})
	require.NoError(t, err)

	require.Equal(t, "Goodbye!", b.String())
}

func TestTemplate_If(t *testing.T) {
	template, err := NewTemplate("hello.html", "{{if name != nil}}Hello!{{else}}Goodbye!{{end}}")
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, nil, map[string]any{"name": "Fox Mulder"})
	require.NoError(t, err)

	require.Equal(t, "Hello!", b.String())

	b = new(bytes.Buffer)
	err = template.Execute(b, nil, map[string]any{})
	require.NoError(t, err)

	require.Equal(t, "Goodbye!", b.String())
}

func TestTemplate_IfTrue(t *testing.T) {
	template, err := NewTemplate("hello.html", "{{if true == true}}Hello!{{end}}")
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, nil, map[string]any{})
	require.NoError(t, err)

	require.Equal(t, "Hello!", b.String())
}

func TestTemplate_IfTruthy(t *testing.T) {
	template, err := NewTemplate("hello.html", "{{if name}}Hello!{{end}}")
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, nil, map[string]any{"name": "Fox Mulder"})
	require.NoError(t, err)

	require.Equal(t, "Hello!", b.String())
}

func TestTemplate_IfFalse(t *testing.T) {
	template, err := NewTemplate("hello.html", "{{if false == false}}Hello!{{end}}")
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, nil, map[string]any{})
	require.NoError(t, err)

	require.Equal(t, "Hello!", b.String())
}

func TestTemplate_IfFalsNoElse(t *testing.T) {
	template, err := NewTemplate("hello.html", "{{if false}}Hello!{{end}}")
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, nil, map[string]any{})
	require.NoError(t, err)

	require.Equal(t, "", b.String())
}

func TestTemplateRange(t *testing.T) {
	template, err := NewTemplate("hello.html", `
	{{range $i, $val in people}}
		<h1>Hello, {{$val}}, person #{{$i}}</h1>
	{{end}}
	`)

	require.NoError(t, err)
	data := map[string]any{"people": []string{"Fox Mulder", "Dana Scully"}}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := `
	
		<h1>Hello, Fox Mulder, person #0</h1>
	
		<h1>Hello, Dana Scully, person #1</h1>
	
	`
	require.Equal(t, expected, b.String())
}

func TestTemplateRange_SingleVariable(t *testing.T) {
	template, err := NewTemplate("hello.html", `
	{{range $val in people}}
		<h1>Hello, {{$val}}</h1>
	{{end}}
	`)

	require.NoError(t, err)
	data := map[string]any{"people": []string{"Fox Mulder", "Dana Scully"}}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := `
	
		<h1>Hello, 0</h1>
	
		<h1>Hello, 1</h1>
	
	`
	require.Equal(t, expected, b.String())
}

func TestTemplateRange_Map(t *testing.T) {
	template, err := NewTemplate("hello.html", `
	{{range $first, $last in people}}
		<h1>Hello, {{$first}} {{$last}}</h1>
	{{end}}
	`)

	require.NoError(t, err)
	data := map[string]any{"people": map[string]string{"Fox": "Mulder", "Dana": "Scully"}}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := `
	
		<h1>Hello, Dana Scully</h1>
	
		<h1>Hello, Fox Mulder</h1>
	
	`
	require.Equal(t, expected, b.String())
}

func TestTemplateRange_NestedStringConditional(t *testing.T) {
	template, err := NewTemplate("hello.html", `
{{range $first, $last in people}}
	{{if $first == "Fox"}}
		Agent {{$first}} {{$last}}
	{{else}}
		Dr. {{$first}} {{$last}}
	{{end}}
{{end}}
	`)

	require.NoError(t, err)
	data := map[string]any{"people": map[string]string{"Fox": "Mulder", "Dana": "Scully"}}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := `

	
		Dr. Dana Scully
	

	
		Agent Fox Mulder
	

	`
	require.Equal(t, expected, b.String())
}

func TestTemplateRange_Numbers(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{if 1000 == 1000}}hello {{1000}}!{{end}}`)

	require.NoError(t, err)
	data := map[string]any{"people": map[string]string{"Fox": "Mulder", "Dana": "Scully"}}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := `hello 1000!`
	require.Equal(t, expected, b.String())
}

func TestTemplate_NegativeLiteral(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{if -1000 == -1000}}hello {{1000}}!{{end}}`)

	require.NoError(t, err)
	data := map[string]any{"people": map[string]string{"Fox": "Mulder", "Dana": "Scully"}}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := `hello 1000!`
	require.Equal(t, expected, b.String())
}

func TestTemplate_NegativeVariable(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{range $i in people}}{{-$i}}!{{end}}`)

	require.NoError(t, err)
	data := map[string]any{"people": []string{"Fox Mulder", "Dana Scully"}}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := `0!-1!`
	require.Equal(t, expected, b.String())
}

func TestTemplate_NegativeVariableNonInt(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{range $i in people}}{{-$i}}!{{end}}`)

	require.NoError(t, err)
	data := map[string]any{"people": map[string]string{"Fox": "Mulder", "Dana": "Scully"}}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.Error(t, err)
	// TODO validate line information is provided
}

func TestTemplate_Subtraction(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{100 - 5}}`)

	require.NoError(t, err)
	data := map[string]any{"people": map[string]string{"Fox": "Mulder", "Dana": "Scully"}}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := "95"
	require.Equal(t, expected, b.String())
}

func TestTemplate_Addition(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{100 + 5}}`)

	require.NoError(t, err)
	data := map[string]any{"people": map[string]string{"Fox": "Mulder", "Dana": "Scully"}}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := "105"
	require.Equal(t, expected, b.String())
}

func TestTemplate_Multiplication(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{100 * 5}}`)

	require.NoError(t, err)
	data := map[string]any{"people": map[string]string{"Fox": "Mulder", "Dana": "Scully"}}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := "500"
	require.Equal(t, expected, b.String())
}

func TestTemplate_Division(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{100 / 5}}`)

	require.NoError(t, err)
	data := map[string]any{"people": map[string]string{"Fox": "Mulder", "Dana": "Scully"}}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := "20"
	require.Equal(t, expected, b.String())
}

func TestTemplate_Modulo(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{100 % 5}}`)

	require.NoError(t, err)
	data := map[string]any{"people": map[string]string{"Fox": "Mulder", "Dana": "Scully"}}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := "0"
	require.Equal(t, expected, b.String())
}

func TestTemplate_Escape(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{userInput}}`, WithEscapeFunc(HTMLEscape))

	require.NoError(t, err)
	data := map[string]any{"userInput": "<h1>Hello!</h1>"}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := "&lt;h1&gt;Hello!&lt;/h1&gt;"
	require.Equal(t, expected, b.String())
}

func TestTemplate_EscapeSafe(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{userInput}}`, WithEscapeFunc(HTMLEscape))

	require.NoError(t, err)
	data := map[string]any{"userInput": Safe("<h1>Hello!</h1>")}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := "<h1>Hello!</h1>"
	require.Equal(t, expected, b.String())
}

type stringerStruct struct {
	value string
}

func (s *stringerStruct) String() string { return s.value }

func TestTemplate_Stringer(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{userInput}}`, WithEscapeFunc(HTMLEscape))

	require.NoError(t, err)
	data := map[string]any{"userInput": &stringerStruct{value: "foo"}}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := "foo"
	require.Equal(t, expected, b.String())
}

func TestTemplate_Call(t *testing.T) {
	f := func() string { return "omg" }
	template, err := NewTemplate("hello.html", `{{foo()}}`, WithEscapeFunc(HTMLEscape), WithHelpers(map[string]any{"foo": f}))

	require.NoError(t, err)
	data := map[string]any{"userInput": &stringerStruct{value: "foo"}}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := "omg"
	require.Equal(t, expected, b.String())
}

func TestTemplate_CallArgs(t *testing.T) {
	f := func(i int) string { return "you are number " + strconv.Itoa(i) }
	template, err := NewTemplate("hello.html", `{{foo(1)}}`, WithEscapeFunc(HTMLEscape), WithHelpers(map[string]any{"foo": f}))

	require.NoError(t, err)
	data := map[string]any{}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := "you are number 1"
	require.Equal(t, expected, b.String())
}

func TestTemplate_CallChain(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{user.Name.Initials()}}`, WithEscapeFunc(HTMLEscape))

	require.NoError(t, err)
	data := map[string]any{"user": user{Name: name{First: "Fox", Last: "Mulder"}}}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := "FM"
	require.Equal(t, expected, b.String())
}

func TestTemplate_CallNestedChain(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{user.GetName().Initials()}}`, WithEscapeFunc(HTMLEscape))

	require.NoError(t, err)
	data := map[string]any{"user": user{Name: name{First: "Fox", Last: "Mulder"}}}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := "FM"
	require.Equal(t, expected, b.String())
}

func TestTemplate_Hash(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{ { foo: 1, bar: 2} }}`, WithEscapeFunc(HTMLEscape))

	require.NoError(t, err)
	data := map[string]any{}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := "map[bar:2 foo:1]"
	require.Equal(t, expected, b.String())
}

func TestTemplate_CallHash(t *testing.T) {
	lenHelper := func(m map[string]any) int {
		return len(m)
	}
	template, err := NewTemplate("hello.html", `{{len({foo: 1, bar: 2})}}`, WithEscapeFunc(HTMLEscape), WithHelpers(map[string]any{"len": lenHelper}))

	require.NoError(t, err)
	data := map[string]any{}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := "2"
	require.Equal(t, expected, b.String())
}

func TestTemplate_BracketAccess(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{ {foo: 1, bar: 2}["foo"] }}`, WithEscapeFunc(HTMLEscape))

	require.NoError(t, err)
	data := map[string]any{}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := "1"
	require.Equal(t, expected, b.String())
}

func TestTemplate_Nil(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{ value }}`)
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, nil, map[string]any{})
	require.NoError(t, err)

	expected := ""
	require.Equal(t, expected, b.String())
}

func TestTemplate_NotFalsy(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{ !value }}`)
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, nil, map[string]any{"value": false})
	require.NoError(t, err)

	expected := "true"
	require.Equal(t, expected, b.String())
}

func TestTemplate_NotTruthy(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{ !value }}`)
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, nil, map[string]any{"value": true})
	require.NoError(t, err)

	expected := "false"
	require.Equal(t, expected, b.String())
}

func TestTemplate_HelperCallError(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{ foo() }}`, WithHelpers(map[string]any{"foo": func(x int) {}}))
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, nil, map[string]any{"value": true})
	require.Error(t, err)
	require.ErrorContains(t, err, "error calling function 'foo'")
	require.ErrorContains(t, err, "too few input arguments")
}

func TestTemplate_IfHelper(t *testing.T) {
	lenHelper := func(v any) int { return reflect.ValueOf(v).Len() }
	template, err := NewTemplate("hello.html", `{{ if len(foo) == 0 }}bar{{end}}`, WithHelpers(map[string]any{"len": lenHelper}))
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, nil, map[string]any{"foo": []string{}})
	require.NoError(t, err)

	require.Equal(t, "bar", b.String())
}

func TestTemplateRange_Channel(t *testing.T) {
	template, err := NewTemplate("hello.html", `
	{{range $i, $val in people}}
		<h1>Hello, {{$val}}, person #{{$i}}</h1>
	{{end}}
	`)
	require.NoError(t, err)

	ch := make(chan string, 2)
	ch <- "Fox Mulder"
	ch <- "Dana Scully"
	data := map[string]any{"people": ch}

	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := `
	
		<h1>Hello, Fox Mulder, person #0</h1>
	
		<h1>Hello, Dana Scully, person #1</h1>
	
	`
	require.Equal(t, expected, b.String())
}

func TestTemplateRange_Array(t *testing.T) {
	template, err := NewTemplate("hello.html", `
	{{range $i, $val in people}}
		<h1>Hello, {{$val}}, person #{{$i}}</h1>
	{{end}}
	`)
	require.NoError(t, err)

	data := map[string]any{"people": [2]string{"Fox Mulder", "Dana Scully"}}

	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := `
	
		<h1>Hello, Fox Mulder, person #0</h1>
	
		<h1>Hello, Dana Scully, person #1</h1>
	
	`
	require.Equal(t, expected, b.String())
}

func TestTemplate_IfWithSubtraction(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{if 999 == 1000 - 1}}hello 999!{{end}}`)

	require.NoError(t, err)
	data := map[string]any{"people": map[string]string{"Fox": "Mulder", "Dana": "Scully"}}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := `hello 999!`
	require.Equal(t, expected, b.String())
}

func TestTemplate_ArrayAccessInt64(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{ foo[i] }}`)
	require.NoError(t, err)

	data := map[string]any{"foo": []string{"bar"}, "i": int64(0)}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := `bar`
	require.Equal(t, expected, b.String())
}

type callableType struct {
	body string
}

func (c *callableType) UpperBody() string {
	return strings.ToUpper(c.body)
}

func TestTemplate_ValueMethods(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{ value.UpperBody() }}`)
	require.NoError(t, err)

	data := map[string]any{"value": &callableType{body: "hello"}}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := `HELLO`
	require.Equal(t, expected, b.String())
}

func TestTemplate_VarGreaterThan(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{ if Page > 1}}foo{{end}}`)
	require.NoError(t, err)

	data := map[string]any{"Page": 2}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := `foo`
	require.Equal(t, expected, b.String())
}

func TestTemplate_VarLessThan(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{ if Page < 1}}foo{{end}}`)
	require.NoError(t, err)

	data := map[string]any{"Page": 0}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := `foo`
	require.Equal(t, expected, b.String())
}

func TestTemplate_VarGreaterThanEqual(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{ if Page >= 1}}foo{{end}}`)
	require.NoError(t, err)

	data := map[string]any{"Page": 1}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := `foo`
	require.Equal(t, expected, b.String())
}

func TestTemplate_VarLessThanEqual(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{ if Page <= 1}}foo{{end}}`)
	require.NoError(t, err)

	data := map[string]any{"Page": 1}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := `foo`
	require.Equal(t, expected, b.String())
}

func TestTemplate_MathOrder(t *testing.T) {
	lenHelper := func(m []string) int {
		return len(m)
	}
	template, err := NewTemplate("hello.html", `{{ if 0 == len(Items) - 1 }}foo{{end}}`, WithHelpers(map[string]any{"len": lenHelper}))
	require.NoError(t, err)

	data := map[string]any{"Items": []string{"foo"}}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)
	require.NoError(t, err)

	expected := `foo`
	require.Equal(t, expected, b.String())
}

func TestTemplate_MissingHelper(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{len(foo)}}`)
	require.NoError(t, err)

	data := map[string]any{"Items": []string{"foo"}}
	b := new(bytes.Buffer)
	err = template.Execute(b, nil, data)

	require.ErrorContains(t, err, "function 'len' not defined")
}

func TestTemplate_MissingMapValue(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{ { Errors: Errors } }}`)
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, nil, map[string]any{})

	require.NoError(t, err)
}

func TestTemplate_MissingMapAccessValue(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{ Foo["bar"] }}`)
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, nil, map[string]any{"Foo": map[string]string{}})

	require.NoError(t, err)
}

func TestTemplate_MapAccessInMap(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{ { Errors: Errors["first"] } }}`)
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, nil, map[string]any{"Errors": map[string]string{"first": "foo"}})

	require.NoError(t, err)
}

func TestTemplate_MapAccessInMap_IntInt64(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{ { Errors: Errors[ID] } }}`)
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(
		b,
		nil,
		map[string]any{
			"Errors": map[int64]string{1: "foo"},
			"ID":     int(1),
		},
	)

	require.NoError(t, err)
}

func TestTemplate_MapAccessInMap_WrongTypes(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{ { Errors: Errors[ID] } }}`)
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(
		b,
		nil,
		map[string]any{
			"Errors": map[string]string{"1": "foo"},
			"ID":     int(1),
		},
	)

	require.ErrorContains(t, err, "cannot access map of type map[string]string with access of type int")
}

func TestTemplate_StringConcat(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{ "Hello, " + Name }}`)
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, nil, map[string]any{"Name": "Fox Mulder"})
	require.NoError(t, err)

	require.Equal(t, "Hello, Fox Mulder", b.String())

}

func TestTemplate_StringConcat_SafeUnsafe(t *testing.T) {
	template, err := NewTemplate("hello.html", `{{ Greeting + Name + "<script>" + Ending}}`)
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(
		b,
		nil,
		map[string]any{
			"Greeting": Safe("<a href=\"#\">"),
			"Name":     "Fox Mulder",
			"Ending":   Safe("</a>"),
		},
	)
	require.NoError(t, err)

	require.Equal(t, `<a href="#">Fox Mulder&lt;script&gt;</a>`, b.String())

}

func TestEngine_Error_Invalid_Maths(t *testing.T) {
	engine := NewEngine(NoEscape)
	err := engine.Register("hello", "{{Age - 1}}")
	b := new(bytes.Buffer)
	err = engine.Render(b, "hello", nil)

	require.Errorf(t, err, "can't subtract invalid from int")
}

func TestInt64_Equal(t *testing.T) {
	engine := NewEngine(NoEscape)
	err := engine.Register("hello", "{{ID == 1}}")
	b := new(bytes.Buffer)
	err = engine.Render(b, "hello", map[string]any{"ID": int64(1)})

	require.NoError(t, err)

	require.Equal(t, `true`, b.String())
}
