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

### Conditionals

Bat supports `if` statements, and the `!=` and `==` operators.

```html
{{if user != nil}}
    <a href="/login">Login</a>
{{else}}
    <a href="/profile">View your profile</a>
{{end}}
```

## TODO

- [ ] Add `each` functionality
- [x] Add `if` and `else` functionality
- [ ] Emit better error messages and validate them with tests
- [ ] Create an engine struct that will enable partials, helper functions, and
      custom escaping functions.
- [ ] Support strings in templates
- [ ] Support numbers
- [ ] Add basic math operations
- [ ] Simple map class `{ "foo": bar }` for use with partials

## Maybe

- Add &&, and || operators for more complex conditionals
- Replace `{{end}}` with named end blocks, like `{{/if}}`
- Add support for `{{else if <expression>}}`
- Support the not operator, e.g. `if !foo`

## Don't

- Add parens for complex options
- Variable declarations that look like provided data access
