# Bat

A mustache like (`{{foo.bar}}`) templating engine for Go. This is still very
much WIP, but contributions and issues are welcome.

## Usage


Given a file, `index.bat`:

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

## TODO

- [ ] Add `each` functionality
- [ ] Add `if` and `else` functionality
- [ ] Emit better error messages and validate them with tests
- [ ] Create engines that will enable partials, helper functions, and custom escaping
