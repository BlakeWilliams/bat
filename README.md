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

Here's an overview of more advanced usage:

### Primitives:

Bat supports the following primitives that can be used within `{{}}`
expressions:

- booleans - `true` and `false`
- nil - `nil`
- strings - `"string value"` and `"string with \"escaped\" values"`
- integers - `1000` and `-1000`

### Conditionals

Bat supports `if` statements, and the `!=` and `==` operators.

```html
{{if user != nil}}
    <a href="/login">Login</a>
{{else}}
    <a href="/profile">View your profile</a>
{{end}}
```

### Iterators

Iteration is supported via the `range` keyword. Both slices and maps are supported.

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

In the example above, range defines two variables which __must__ begin with a $
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

If a map is passed to `range`, it will attempt to sort it before iteration if the key is able to be compared and is implemented in the `internal/mapsort` package.


## TODO

- [x] Add `each` functionality (see the section on `range`)
- [x] Add `if` and `else` functionality
- [ ] Emit better error messages and validate them with tests
- [ ] Create an engine struct that will enable partials, helper functions, and
      custom escaping functions.
- [x] Support strings in templates
- [x] Support integer numbers
- [ ] Add basic math operations
- [ ] Simple map class `{ "foo": bar }` for use with partials
- [ ] Improve stringify logic in the executor (bat.go)
- [ ] Support channels in `range`
- [ ] Trim whitespace by default, add control characters to avoid trimming.

## Maybe

- Add &&, and || operators for more complex conditionals
- Replace `{{end}}` with named end blocks, like `{{/if}}`
- Add support for `{{else if <expression>}}`
- Support the not operator, e.g. `if !foo`
- Track and error on undefined variable usage in the parsing stage

## Don't

- Add parens for complex options
- Variable declarations that look like provided data access (use $ for template locals, plain identifiers for everything else)
