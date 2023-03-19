package bat

import (
	"bytes"
	"embed"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEngine(t *testing.T) {
	engine := NewEngine(NoEscape)

	engine.Helper("omg", func() string {
		return "omg"
	})
	err := engine.Register("foo", "{{omg()}}")
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = engine.Render(b, "foo", map[string]any{})
	require.NoError(t, err)

	require.Equal(t, "omg", b.String())
}

//go:embed fixtures
var fixtures embed.FS

func TestEngine_AutoRegister(t *testing.T) {
	engine := NewEngine(NoEscape)

	err := engine.AutoRegister(fixtures, "", ".html")
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = engine.Render(b, "fixtures/home.html", map[string]any{"siteName": "bat"})
	require.NoError(t, err)

	require.Equal(t, "<h1>Welcome to bat</h1>\n", b.String())

	b = new(bytes.Buffer)
	err = engine.Render(b, "fixtures/users/hello.html", map[string]any{"name": "Fox"})
	require.NoError(t, err)

	require.Equal(t, "<h1>Hello Fox</h1>\n", b.String())
}

func TestEngine_EscapesHTML(t *testing.T) {
	engine := NewEngine(HTMLEscape)

	err := engine.Register("foo", "{{\"<h1>hi</h1>\"}}")
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = engine.Render(b, "foo", map[string]any{})
	require.NoError(t, err)

	require.Equal(t, "<h1>hi</h1>", b.String())
}

func TestEngine_DefaultHelper_Safe(t *testing.T) {
	engine := NewEngine(NoEscape)

	err := engine.Register("foo", "{{safe(\"<h1>hi</h1>\")}}")
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = engine.Render(b, "foo", map[string]any{})
	require.NoError(t, err)

	require.Equal(t, "<h1>hi</h1>", b.String())
}

func TestEngine_DefaultHelper_Len(t *testing.T) {
	engine := NewEngine(NoEscape)

	err := engine.Register("foo", `{{len("some value")}}`)
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = engine.Render(b, "foo", map[string]any{})
	require.NoError(t, err)

	require.Equal(t, "10", b.String())
}

func TestEngine_DefaultHelper_Partial(t *testing.T) {
	engine := NewEngine(NoEscape)

	err := engine.Register("hello", "{{name}}")
	require.NoError(t, err)
	err = engine.Register("foo", `Hi {{partial("hello", {name: name})}}`)
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = engine.Render(b, "foo", map[string]any{"name": "Fox Mulder"})
	require.NoError(t, err)

	require.Equal(t, "Hi Fox Mulder", b.String())
}

func TestEngine_Errors(t *testing.T) {
	engine := NewEngine(NoEscape)

	err := engine.Register("hello", "{{name[0]}}")
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = engine.Render(b, "hello", map[string]any{"name": "Fox Mulder"})
	require.Error(t, err)
	require.ErrorContains(t, err, "starting on line 1")
	require.ErrorContains(t, err, "{{name[0]}}")
}

func TestEngine_Render_Layout(t *testing.T) {
	engine := NewEngine(NoEscape)

	err := engine.Register("layout", `<h1>HELLO {{ ChildContent }}!</h1>`)
	require.NoError(t, err)
	err = engine.Register("hello", `{{ layout("layout") }}{{ name }}`)
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = engine.Render(b, "hello", map[string]any{"name": "Fox Mulder"})
	require.NoError(t, err)

	require.Equal(t, "<h1>HELLO Fox Mulder!</h1>", b.String())
}

func TestEngine_Render_Nested_Layout(t *testing.T) {
	engine := NewEngine(NoEscape)

	err := engine.Register("root", "<html>{{ ChildContent }}</html>")
	require.NoError(t, err)
	err = engine.Register("layout", `{{ layout("root") }}<h1>HELLO {{ ChildContent }}!</h1>`)
	require.NoError(t, err)
	err = engine.Register("hello", `{{ layout("layout") }}{{ name }}`)
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = engine.Render(b, "hello", map[string]any{"name": "Fox Mulder"})
	require.NoError(t, err)

	require.Equal(t, "<html><h1>HELLO Fox Mulder!</h1></html>", b.String())
}

func TestEngine_Render_Layout_MultipleCalls(t *testing.T) {
	engine := NewEngine(NoEscape)

	err := engine.Register("layout", `<h1>HELLO {{ ChildContent }}!</h1>`)
	require.NoError(t, err)
	err = engine.Register("hello", `{{ layout("layout") }}{{ layout("layout") }}`)
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = engine.Render(b, "hello", map[string]any{"name": "Fox Mulder"})
	require.ErrorContains(t, err, "layout already set")
}

func TestEngine_Render_Layout_Missing(t *testing.T) {
	engine := NewEngine(NoEscape)
	err := engine.Register("hello", `{{ layout("layout") }}`)
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = engine.Render(b, "hello", map[string]any{"name": "Fox Mulder"})
	require.Error(t, err)
}

func TestEngine_Render_Layout_InheritsData(t *testing.T) {
	engine := NewEngine(NoEscape)
	err := engine.Register("layout", `<{{ Tag }}>HELLO {{ ChildContent }}!</{{ Tag }}>`)
	require.NoError(t, err)
	err = engine.Register("hello", `{{ layout("layout") }}{{ name }}`)
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = engine.Render(b, "hello", map[string]any{"name": "Fox Mulder", "Tag": "h2"})
	require.NoError(t, err)
	require.Equal(t, "<h2>HELLO Fox Mulder!</h2>", b.String())
}

func TestEngine_Render_Nested_LocalHelper(t *testing.T) {
	engine := NewEngine(NoEscape)

	err := engine.Register("root", "<html>{{ ChildContent }}</html>")
	require.NoError(t, err)
	err = engine.Register("layout", `{{ layout("root") }}<h1>HELLO {{ ChildContent }}!</h1>`)
	require.NoError(t, err)
	err = engine.Register("hello", `{{ layout("layout") }}{{ omg() }}`)
	require.NoError(t, err)

	helpers := map[string]any{
		"omg": func() string { return "omg" },
	}
	b := new(bytes.Buffer)
	err = engine.RenderWithHelpers(b, "hello", helpers, map[string]any{"name": "Fox Mulder"})
	require.NoError(t, err)

	require.Equal(t, "<html><h1>HELLO omg!</h1></html>", b.String())
}

func TestEngine_DefaultHelper_Partial_Helpers(t *testing.T) {
	engine := NewEngine(NoEscape)

	err := engine.Register("hello", "{{name}}. {{ omg() }}")
	require.NoError(t, err)
	err = engine.Register("foo", `Hi {{partial("hello", {name: name})}}`)
	require.NoError(t, err)

	helpers := map[string]any{
		"omg": func() string { return "omg" },
	}

	b := new(bytes.Buffer)
	err = engine.RenderWithHelpers(b, "foo", helpers, map[string]any{"name": "Fox Mulder"})
	require.NoError(t, err)

	require.Equal(t, "Hi Fox Mulder. omg", b.String())
}
