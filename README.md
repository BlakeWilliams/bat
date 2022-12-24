# Bat

A mustache like (`{{foo.bar}}`) templating engine for Go. This is still very
much WIP, but contributions and issues are welcome.

## Usage

Given a file, `index.batml`:

```
<h1>Hello {{Team.Name}}</h1>
```

Create a new template and execute it:

```go
content, _ := ioutil.ReadFile("index.bat")
bat.NewTemplate(content)

t := team{
    Name: "Foo",
}
bat.Execute(map[string]any{"Team": team})
```

### Engine

Bat provides an engine that allows you to register templates and provides
default, as well as user provided helper functions to those templates.

```go
engine := bat.NewEngine(bat.HTMLEscape)
engine.Register("index.bat", "<h1>Hello {{Team.Name}}</h1>")
```

or, you can use `AutoRegister` to automatically register all templates in a
directory. This is useful with the Go embed package:

```go
//go:embed templates
var templates embed.FS

engine := bat.NewEngine(bat.HTMLEscape)
engine.AutoRegister(templates, ".html")

engine.Render("templates/users/signup.html", map[string]any{"Team": team})
```

#### Built-in helpers

- `safe` - marks a value as safe to be rendered. This is useful for rendering
  HTML. For example, `{{safe("<h1>Foo</h1>")}}` will render `<h1>Foo</h1>`.
- `len` - returns the length of a slice or map. For example, `{{len(Users)}}` will
  return the length of the `Users` slice.
- `partial` - renders a partial template. For example, `{{partial("header", {foo: "bar"})}}`
  will render the `header` template with the provided map as locals.

Here's an overview of more advanced usage:

### Primitives

Bat supports the following primitives that can be used within `{{}}`
expressions:

- booleans - `true` and `false`
- nil - `nil`
- strings - `"string value"` and `"string with \"escaped\" values"`
- integers - `1000` and `-1000`
- maps - `{ foo: 1, bar: "two" }`

### Data Access

Templates accept data in the form of `map[string]any`. The strings must be
valid identifiers in order to be access, which start with an alphabetical
character following by any number of alphanumerical characters.

The template `{{userName}}` would attempt to access the `userName` key from the
provided data map.

e.g.

```go
t := bat.NewTemplate(`{{userName}}!`)
out := new(bytes.Buffer)

// outputs "gogopher!"
t.Execute(out, map[string]{"Username": "gogopher"}
```

Chaining and method calls are also supported:

```go
type Name struct {
    First string
    Last string
}

type User struct {
    Name Name
}

func (n Name) Initials() string {
    return n.First[0:1] + n.Last[0:1]
}

t := bat.NewTemplate(`{{user.Name.Initials()}}!`)
out := new(bytes.Buffer)

user := User{
    Name: Name{
        First: "Fox",
        Last: "Mulder",
    }
}

// outputs "FM!"
t.Execute(out, map[string]{"user": user}
```

Finally, map/slice/array access is supported via `[]`:

```html
<h1>{{user[0].Name.First}}</h1>
```

### Conditionals

Bat supports `if` statements, and the `!=` and `==` operators.

```html
{{if user != nil}}
<a href="/login">Login</a>
{{else}}
<a href="/profile">View your profile</a>
{{end}}
```

### Not

The `!` operator can be used to negate an expression and return a boolean

```html
{{!true}}
```

The above will render `false`.

### Iterators

Iteration is supported via the `range` keyword. Supported types are slices, maps, arrays, and channels.

```html
{{range $index, $name in data}}
<h1>Hello {{$name}}, number {{$index}}</h1>
{{end}}
```

Given `data` being defined as: `[]string{"Fox Mulder", "Dana Scully"}`, the resulting output would look like:

```html
<h1>Hello Fox Mulder, number 0</h1>

<h1>Hello Dana Scully, number 1</h1>
```

In the example above, range defines two variables which **must** begin with a $
so they don't conflict with `data` passed into the template.

The `range` keyword can also be used with a single variable, providing only the
key or index to the iterator:

```html
{{range $index in data}}
<h1>Hello person {{$index}}</h1>
{{end}}
```

Given `data` being defined as: `[]string{"Fox Mulder", "Dana Scully"}`, the resulting output would look like:

```html
<h1>Hello person 0</h1>

<h1>Hello person 1</h1>
```

If a map is passed to `range`, it will attempt to sort it before iteration if
the key is able to be compared and is implemented in the `internal/mapsort`
package.

### Helper functions

Helper functions can be provided directly to templates using the `WithHelpers` function when instantiating a template.

e.g.

```
helloHelper := func(name string) string {
    return fmt.Sprintf("Hello %s!", name)
}

t := bat.NewTemplate(`{{hello "there"}}`, WithHelpers(map[string]any{"hello": helloHelper}))

// output "Hello there!"
out := new(bytes.Buffer)
t.Execute(out, map[string]any{})
```

### Escaping

Templates can be provided a custom escape function with the signature
`func(string) string` that will be called on the resulting output from `{{}}`
blocks.

There are two escape functions that can be utilized, `NoEscape` which does no
escaping, and `HTMLEscape` which delegates to `html.EscapeString`, which
escapes HTML.

The default escape function is `HTMLEscape` for safety reasons.

e.g.

```go
// This template will escape HTML from the output of `{{}}` blocks
t := NewTemplate("{{foo}}", WithEscapeFunc(HTMLEscape))
```

Escaping can be avoided by returning the `bat.Safe` type from the result of a
`{{}}` block.

e.g.

```go
t := bat.NewTemplate(`{{output}}`, WithEscapeFunc(HTMLEscape))

// output "Hello there!"
out := new(bytes.Buffer)

// outputs &lt;h1&gt;Hello!&lt;/h1^gt;
t.Execute(out, map[string]any{"output": "<h1>Hello!</h1>"})

// outputs <h1>Hello!</h1>
t.Execute(out, map[string]any{"output": bat.Safe("<h1>Hello!</h1>")})
```

### Math

Basic math is supported, with some caveats. When performing math operations,
the left most type is converted into the right most type, when possible:

```go
// int32 - int64
   100   -   200 // returns int64
```

The following operations are supported:

- `-` Subtraction
- `+` Addition
- `*` Multiplication
- `/` Division
- `%` Modulus

More comprehensive casting logic would be welcome in the form of a PR.

## TODO

- [x] Add `each` functionality (see the section on `range`)
- [x] Add `if` and `else` functionality
- [x] Emit better error messages and validate them with tests (template execution)
- [ ] Emit better error messages from lexer and parser
- [x] Create an engine struct that will enable partials, helper functions, and
      custom escaping functions.
- [x] Add escaping support to templates
- [x] Support strings in templates
- [x] Support integer numbers
- [x] Add basic math operations
- [x] Simple map class `{ "foo": bar }` for use with partials
- [ ] Improve stringify logic in the executor (`bat.go`)
- [x] Support channels in `range`
- [ ] Trim whitespace by default, add control characters to avoid trimming.
- [x] Support method calls
- [x] Support helpers
- [x] Support map/slice array access `[]`

## Maybe

- Add &&, and || operators for more complex conditionals
- ~Replace `{{end}}` with named end blocks, like `{{/if}}`~ rejected
- Add support for `{{else if <expression>}}`
- ~Support the not operator, e.g. `if !foo`~ done
- Track and error on undefined variable usage in the parsing stage

## Don't

- Add parens for complex options
- Variable declarations that look like provided data access (use $ for template locals, plain identifiers for everything else)
- Add string concatenation
