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

	err := engine.AutoRegister(fixtures, ".html")
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

	err := engine.Register("foo", "{{<h1>hi</h1>\"}}")
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
