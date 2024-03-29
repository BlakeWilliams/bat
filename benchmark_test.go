package bat

import (
	"bytes"
	"html/template"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkHelloWorld(b *testing.B) {
	batTemplate, err := NewTemplate("hello.html", `Hello {{name}}`, WithEscapeFunc(HTMLEscape))
	require.NoError(b, err)

	htmlTemplate, err := template.New("foo").Parse(`Hello {{.Name}}`)
	require.NoError(b, err)

	args := map[string]any{"name": "world"}

	batOutput := new(bytes.Buffer)
	batTemplate.Execute(batOutput, nil, args)

	htmlOutput := new(bytes.Buffer)
	batTemplate.Execute(htmlOutput, nil, args)

	require.Equal(b, batOutput.String(), htmlOutput.String())

	b.Run("bat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			batTemplate.Execute(io.Discard, nil, args)
		}
	})

	b.Run("template/html", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			htmlTemplate.Execute(io.Discard, args)
		}
	})

}

func BenchmarkRangeIf(b *testing.B) {
	batTemplate, err := NewTemplate("hello.html", `{{range $_, $name in Names}}{{if $name != "Smoking Man"}}Hello {{$name}}{{else}}Ugh, {{$name}}{{end}}{{end}}`, WithEscapeFunc(HTMLEscape))
	require.NoError(b, err)

	htmlTemplate, err := template.New("foo").Parse(`{{range $name := .Names}}{{if ne $name "Smoking Man"}}Hello {{$name}}{{else}}Ugh, {{$name}}{{end}}{{end}}`)
	require.NoError(b, err)

	args := map[string]any{"Names": []string{"Fox", "Dana", "Smoking Man"}}

	batOutput := new(bytes.Buffer)
	batTemplate.Execute(batOutput, nil, args)

	htmlOutput := new(bytes.Buffer)
	batTemplate.Execute(htmlOutput, nil, args)

	require.Equal(b, batOutput.String(), htmlOutput.String())

	b.Run("bat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			batTemplate.Execute(io.Discard, nil, args)
		}
	})

	b.Run("template/html", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			htmlTemplate.Execute(io.Discard, args)
		}
	})

}
