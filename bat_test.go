package bat

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemplate(t *testing.T) {
	template, err := NewTemplate("<h1>Hello {{name}}</h1>")
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, map[string]any{"name": "Fox Mulder"})
	require.NoError(t, err)

	require.Equal(t, "<h1>Hello Fox Mulder</h1>", b.String())
}

type user struct {
	Name name
}

type name struct {
	First string
	Last  string
}

func TestTemplateDots(t *testing.T) {
	user := user{
		Name: name{First: "Fox", Last: "Mulder"},
	}

	template, err := NewTemplate("<h1>Hello {{user.Name.First}} {{user.Name.Last}}</h1>")
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, map[string]any{"user": user})
	require.NoError(t, err)

	require.Equal(t, "<h1>Hello Fox Mulder</h1>", b.String())

}

func TestTemplateDots_Map(t *testing.T) {
	user := map[string]map[string]string{
		"user": {
			"name": "Fox Mulder",
		},
	}

	template, err := NewTemplate("<h1>Hello {{details.user.name}}</h1>")
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, map[string]any{"details": user})
	require.NoError(t, err)

	require.Equal(t, "<h1>Hello Fox Mulder</h1>", b.String())
}

func TestTemplateDotsNil(t *testing.T) {
	template, err := NewTemplate("<h1>Hello {{details.user.name}}</h1>")
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, map[string]any{})
	require.Error(t, err)
	require.ErrorContains(t, err, "attempted to access property `user` on nil value")
	require.ErrorContains(t, err, "on line 1")
}

func TestTemplate_If(t *testing.T) {
	template, err := NewTemplate("{{if name != nil}}Hello!{{else}}Goodbye!{{end}}")
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, map[string]any{"name": "Fox Mulder"})
	require.NoError(t, err)

	require.Equal(t, "Hello!", b.String())

	b = new(bytes.Buffer)
	err = template.Execute(b, map[string]any{})
	require.NoError(t, err)

	require.Equal(t, "Goodbye!", b.String())
}

func TestTemplate_IfTrue(t *testing.T) {
	template, err := NewTemplate("{{if true == true}}Hello!{{end}}")
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, map[string]any{})
	require.NoError(t, err)

	require.Equal(t, "Hello!", b.String())
}

func TestTemplate_IfFalse(t *testing.T) {
	template, err := NewTemplate("{{if false == false}}Hello!{{end}}")
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = template.Execute(b, map[string]any{})
	require.NoError(t, err)

	require.Equal(t, "Hello!", b.String())
}

func TestTemplateRange(t *testing.T) {
	template, err := NewTemplate(`
	{{range $i, $val in people}}
		<h1>Hello, {{$val}}, person #{{$i}}</h1>
	{{end}}
	`)

	require.NoError(t, err)
	data := map[string]any{"people": []string{"Fox Mulder", "Dana Scully"}}
	b := new(bytes.Buffer)
	err = template.Execute(b, data)
	require.NoError(t, err)

	expected := `
	
		<h1>Hello, Fox Mulder, person #0</h1>
	
		<h1>Hello, Dana Scully, person #1</h1>
	
	`
	require.Equal(t, expected, b.String())
}

func TestTemplateRange_SingleVariable(t *testing.T) {
	template, err := NewTemplate(`
	{{range $val in people}}
		<h1>Hello, {{$val}}</h1>
	{{end}}
	`)

	require.NoError(t, err)
	data := map[string]any{"people": []string{"Fox Mulder", "Dana Scully"}}
	b := new(bytes.Buffer)
	err = template.Execute(b, data)
	require.NoError(t, err)

	expected := `
	
		<h1>Hello, 0</h1>
	
		<h1>Hello, 1</h1>
	
	`
	require.Equal(t, expected, b.String())
}

func TestTemplateRange_Map(t *testing.T) {
	template, err := NewTemplate(`
	{{range $first, $last in people}}
		<h1>Hello, {{$first}} {{$last}}</h1>
	{{end}}
	`)

	require.NoError(t, err)
	data := map[string]any{"people": map[string]string{"Fox": "Mulder", "Dana": "Scully"}}
	b := new(bytes.Buffer)
	err = template.Execute(b, data)
	require.NoError(t, err)

	expected := `
	
		<h1>Hello, Dana Scully</h1>
	
		<h1>Hello, Fox Mulder</h1>
	
	`
	require.Equal(t, expected, b.String())
}

func TestTemplateRange_NestedStringConditional(t *testing.T) {
	template, err := NewTemplate(`
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
	err = template.Execute(b, data)
	require.NoError(t, err)

	expected := `

	
		Dr. Dana Scully
	

	
		Agent Fox Mulder
	

	`
	require.Equal(t, expected, b.String())
}
