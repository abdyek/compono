# Compono

Compono is a **platform-agnostic**, component-based domain-specific language (DSL) that extends Markdown syntax with reusable components.

Originally developed for [Umono CMS](https://github.com/umono-cms/umono), Compono can be used in any Go project that needs a flexible templating solution.

## Installation

```bash
go get github.com/umono-cms/compono
```

## Quick Start

```go
package main

import (
    "bytes"
    "fmt"
    "github.com/umono-cms/compono"
)

func main() {
    c := compono.New()

    source := []byte(`{{ SAY_HELLO name="World" }}

~ SAY_HELLO name="Guest"
# Hello, {{ name }}!
`)

    var buf bytes.Buffer
    if err := c.Convert(source, &buf); err != nil {
        panic(err)
    }

    fmt.Println(buf.String())
    // Output: <h1>Hello, World!</h1>
}
```

## Syntax

### Markdown Support

Compono supports common Markdown elements:

```
# Heading 1
## Heading 2
### Heading 3

This is a paragraph with **bold** and *italic* text.

`inline code`

[Link text](https://example.com)
```

Code blocks are also supported:

~~~
```go
fmt.Println("Hello")
```
~~~

### Components

Components are the core feature of Compono. They allow you to create reusable content blocks.

#### Defining a Local Component

Local components are defined in the same scope where they're used:

```
{{ GREETING }}

~ GREETING
Welcome to our website!
```

The `~ COMPONENT_NAME` syntax marks the beginning of a local component definition. Everything after it becomes the component's content. A component definition ends when another component definition starts or at EOF.

#### Components with Parameters

Components can accept parameters with default values:

```
{{ USER_CARD name="Anonymous" role="Guest" }}

~ USER_CARD name="" role=""
## {{ name }}
*{{ role }}*
```

Parameter types supported:
- Strings: `name="John"`
- Numbers: `age=25`
- Booleans: `active=true`

#### Block vs Inline Components

Components containing multiple paragraphs or block elements are **block components**:

```
{{ ARTICLE }}

~ ARTICLE
# Title
First paragraph.

Second paragraph.
```

Components with single-line content can be used **inline**:

```
Welcome, {{ USERNAME }}!

~ USERNAME
John
```

### Global Components

Global components can be registered once and used across multiple conversions:

```go
c := compono.New()

// Register a global component
c.RegisterGlobalComponent("FOOTER", []byte(`© 2026 My Company`))

// Use it in any conversion
c.Convert([]byte(`
# Page Title
Content here...
{{ FOOTER }}
`), &buf)
```

Global components can also have parameters:

```go
c.RegisterGlobalComponent("BLOG_PAGE", []byte(`title="" content=""
## {{ title }}
{{ content }}`))
```

## Built-in Components

### LINK

Creates an anchor element with optional target blank:

```
{{ LINK text="Visit us" url="https://example.com" new-tab=true }}
```

Output:
```html
<a href="https://example.com" target="_blank" rel="noopener noreferrer">Visit us</a>
```

## Parameters

Components can accept parameters. Each parameter must have a **default value** defined in the component definition.

If a parameter value is not provided during the call, the **default value is used**.

```
{{ SAY_HELLO name="Jane" }}

~ SAY_HELLO name="John"
# Hello, {{ name }}!
```

### Supported Types

Supported parameter types:

- **String** → `name="John"`
- **Number** → `age=25`
- **Bool** → `active=true`
- **Component** → another component can be passed as a parameter

---

### Passing Parameters to Other Components

A parameter can be passed directly to another component call.

```
{{ USER age=31 }}

~ USER age=18
{{ ANOTHER_COMP another-number-param=age }}

~ ANOTHER_COMP another-number-param=0
Number: *{{ another-number-param }}*
```

Here:

- `USER` receives `age`
- it forwards that value to `ANOTHER_COMP`

---

### Passing Components as Parameters

Components themselves can also be passed as parameters.

```
{{ USER name="Yunus Emre" age=31 age-wrapper=AGE_WRAPPER_2 }}

~ USER name="John" age=25 age-wrapper=AGE_WRAPPER_1
# Welcome **{{ name }}**!
{{ age-wrapper age=age }}

~ AGE_WRAPPER_1 age=0
Your age: *{{ age }}*

~ AGE_WRAPPER_2 age=0
*{{ age }}*
```

Here:

- `age-wrapper` receives a **component**
- that component is executed inside `USER`

---

### Global Parameter Visibility in Local Components

When a **global component** defines parameters, those parameters are **visible to local components inside it**.

```
c.RegisterGlobalComponent("PROFILE_PAGE", []byte(`
name="Guest"

{{ PROFILE_CARD }}

~ PROFILE_CARD
## {{ name }}
Welcome to the profile page.
`))
```

Usage:

```
{{ PROFILE_PAGE name="Yunus" }}
```

Output:

```
<h2>Yunus</h2>
<p>Welcome to the profile page.</p>
```

The local component `PROFILE_CARD` can directly access the global parameter `name`.

---

## Error Handling

Compono provides error feedback by rendering placeholders where errors occur.
Fatal errors during conversion stop the process and no output is produced.

## API Reference

### Core Methods

```go
// Create a new Compono instance
c := compono.New()

// Convert source to HTML
err := c.Convert(source []byte, writer io.Writer)

// Register a global component
err := c.RegisterGlobalComponent(name string, source []byte)

// Unregister a global component
err := c.UnregisterGlobalComponent(name string)

// Convert and preview a global component
err := c.ConvertGlobalComponent(name string, source []byte, writer io.Writer)
```

## Component Naming Convention

Component names must be in `SCREAMING_SNAKE_CASE`:

- ✓ `HEADER`
- ✓ `USER_PROFILE`
- ✓ `NAV_MENU_ITEM`
- ✗ `header`
- ✗ `userProfile`

## Parameter Naming Convention

Parameter names must be in `kebab-case`:

- ✓ `name`
- ✓ `user-name`
- ✓ `is-active`
- ✗ `userName`
- ✗ `user_name`

## Component Override Behavior

When multiple components share the same name, Compono follows a clear override hierarchy:

```
Local Component > Global Component > Built-in Component
```

**Local always wins:**

```
{{ LINK }}

~ LINK
I override the built-in LINK component!
```

This outputs `<p>I override the built-in LINK component!</p>` instead of an anchor tag.

**Global overrides built-in:**

```go
c.RegisterGlobalComponent("LINK", []byte(`Custom link behavior`))
```

Now all `{{ LINK }}` calls will use your global definition instead of the built-in one.

This allows you to customize or extend built-in components without modifying the library.

## License

MIT License - see [LICENSE](LICENSE) for details.
