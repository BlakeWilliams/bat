package bat

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
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
		"len": func(v any) int {
			return reflect.ValueOf(v).Len()
		},
		"safe": func(s string) Safe {
			return Safe(s)
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
	t, err := NewTemplate(name, input, WithEscapeFunc(e.escapeFunc), WithHelpers(e.helpers))

	if err != nil {
		return err
	}

	e.templates[name] = t

	return nil
}

// Registers a new template using the given name. Typically name's will be
// relative file paths. e.g. users/new.batml
func (e *Engine) RegisterFile(name string, input string) error {
	t, err := NewTemplate(name, input, WithEscapeFunc(e.escapeFunc), WithHelpers(e.helpers))

	if err != nil {
		return err
	}

	e.templates[name] = t

	return nil
}

// Renders the template with the given name and data to the provider writer.
func (e *Engine) Render(w io.Writer, name string, data map[string]any) error {
	return e.RenderWithHelpers(w, name, nil, data)
}

func (e *Engine) RenderWithHelpers(w io.Writer, name string, helpers map[string]any, data map[string]any) error {
	var layoutName string
	var layoutArgs map[string]any
	if helpers == nil {
		helpers = make(map[string]any, 1)
	}

	helpers["layout"] = func(name string) {
		if layoutName != "" {
			panic("layout already set")
		}

		layoutName = name
	}

	helpers["partial"] = func(name string, data map[string]any) Safe {
		out := new(bytes.Buffer)
		err := e.RenderWithHelpers(out, name, helpers, data)

		if err != nil {
			panic(err)
		}

		return Safe(out.String())
	}

	template, ok := e.templates[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}

	var b bytes.Buffer
	err := template.Execute(&b, helpers, data)
	if err != nil {
		return err
	}

	if layoutName == "" {
		_, _ = w.Write(b.Bytes())
		return err
	}

	layoutData := make(map[string]any, len(data)+1)
	for k, v := range data {
		layoutData[k] = v
	}
	for k, v := range layoutArgs {
		layoutData[k] = v
	}

	layoutData["ChildContent"] = Safe(b.String())

	var tb bytes.Buffer
	err = e.RenderWithHelpers(&tb, layoutName, helpers, layoutData)
	if err != nil {
		return err
	}

	_, _ = w.Write(tb.Bytes())

	return nil
}

// AutoRegister recursivly finds all files with the given extension and
// registers them as a template on the engine. If removePathPrefix is provided,
// it will register templates without the given prefix.
//
// e.g. e.AutoRegister("./templates", ".html") and a file
// ./templates/users/hello.html will register the template with a name of
// "./templates/users/hello.html"
//
// This is designed to be used with the embed package, allowing templates to be
// compiled into the resulting binary.
func (e *Engine) AutoRegister(dir fs.FS, pathPrefix string, extension string) error {
	if pathPrefix != "" && !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix += "/"
	}

	err := fs.WalkDir(dir, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking directory: %s", err)
		}

		if d.IsDir() || !strings.HasSuffix(path, extension) {
			return nil
		}

		f, err := dir.Open(path)

		if err != nil {
			return fmt.Errorf("error opening file: %s", err)
		}

		contents, err := io.ReadAll(f)
		if err != nil {
			return fmt.Errorf("error reading file: %s", err)
		}

		friendlyName := strings.TrimPrefix(path, pathPrefix)
		err = e.Register(friendlyName, string(contents))

		if err != nil {
			return fmt.Errorf("could not register template %s: %w", friendlyName, err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("could not auto-register templates: %w", err)
	}

	return nil
}
