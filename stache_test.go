package stache

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemplate(t *testing.T) {
	template := New("<h1>Hello {{name}}</h1>")

	b := new(bytes.Buffer)
	err := template.Execute(b, map[string]any{"name": "Fox Mulder"})
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

	template := New("<h1>Hello {{user.Name.First}} {{user.Name.Last}}</h1>")

	b := new(bytes.Buffer)
	err := template.Execute(b, map[string]any{"user": user})
	require.NoError(t, err)

	require.Equal(t, "<h1>Hello Fox Mulder</h1>", b.String())

}

func TestTemplateDots_Map(t *testing.T) {
	user := map[string]map[string]string{
		"user": {
			"name": "Fox Mulder",
		},
	}

	template := New("<h1>Hello {{details.user.name}}</h1>")

	b := new(bytes.Buffer)
	err := template.Execute(b, map[string]any{"details": user})
	require.NoError(t, err)

	require.Equal(t, "<h1>Hello Fox Mulder</h1>", b.String())
}
