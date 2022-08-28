package bat

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"reflect"
	"strings"
)

// An Engine represents a collection of templates and helper functions. This
// allows templates to utilize partials and custom escape functions. For most
// applications, there should be 1 engine per-filetype.
type Engine struct {
	templates  map[string]Template
	escapeFunc func(string) string
	helpers    map[string]any
}

// Returns a new engine. NewEngine accepts an escape function that accepts
// un-escpaed text and returns escaped text safe for output.
func NewEngine(escapeFunc func(text string) string) *Engine {
	engine := &Engine{
		escapeFunc: escapeFunc,
		templates:  make(map[string]Template),
	}

	defaultHelpers := map[string]any{
		"safe": func(s string) Safe {
			return Safe(s)
		},
		"partial": func(name string, data map[string]any) string {
			out := new(bytes.Buffer)
			engine.Render(out, name, data)

			return out.String()
		},
	}

	engine.helpers = defaultHelpers

	return engine
}

// Helper declares a new helper function available to templates by using the
// provided name.
//
// If the provided value is not a function this method will panic.
func (e *Engine) Helper(name string, fn any) {
	if reflect.ValueOf(fn).Kind() != reflect.Func {
		panic("provided value must be a function")
	}

	e.helpers[name] = fn
}

// Registers a new template using the given name. Typically name's will be
// relative file paths. e.g. users/new.batml
func (e *Engine) Register(name string, input string) error {
	t, err := NewTemplate(input, WithEscapeFunc(e.escapeFunc), WithHelpers(e.helpers))

	if err != nil {
		return err
	}

	e.templates[name] = t

	return nil
}

// Registers a new template using the given name. Typically name's will be
// relative file paths. e.g. users/new.batml
func (e *Engine) RegisterFile(name string, input string) error {
	t, err := NewTemplate(input, WithEscapeFunc(e.escapeFunc), WithHelpers(e.helpers))

	if err != nil {
		return err
	}

	e.templates[name] = t

	return nil
}

// Renders the template with the given name and data to the provider writer.
func (e *Engine) Render(b io.Writer, name string, data map[string]any) error {
	if template, ok := e.templates[name]; ok {
		return template.Execute(b, data)
	}

	return fmt.Errorf("template %s not found", name)
}

// AutoRegister recursivly finds all files with the given extension and
// registers them as a template on the engine.
//
// e.g. e.AutoRegister("./templates", ".html") and a file
// ./templates/users/hello.html will register the template with a name of
// "./templates/users/hello.html"
//
// This is designed to be used with the embed package, allowing templates to be
// compiled into the resulting binary.
func (e *Engine) AutoRegister(dir fs.FS, extension string) error {
	err := fs.WalkDir(dir, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking directory: %s", err)
		}

		if d.IsDir() || !strings.HasSuffix(path, extension) {
			return nil
		}

		f, err := os.Open(path)

		if err != nil {
			return fmt.Errorf("error opening file: %s", err)
		}

		contents, err := io.ReadAll(f)
		if err != nil {
			return fmt.Errorf("error reading file: %s", err)
		}

		e.Register(path, string(contents))
		return nil
	})

	if err != nil {
		return fmt.Errorf("could not auto-register templates: %s", err)
	}

	return nil
}
