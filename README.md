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

### Global Component Scopes

All global components share one namespace, so two unrelated groups of components can clash: a name may be defined twice, or a call inside one group may resolve to a component of the other. Scopes solve this. A `WithGlobalComponent` can take other `WithGlobalComponent`s as options. These are its **sub components**, and it is their **owner**:

```go
c.Convert(pageSource, w,
	compono.WithGlobalComponent("CARD", cardSource),
	compono.WithGlobalComponent("LAYOUT", layoutSource,
		compono.WithIsolatedScope(),
		compono.WithGlobalComponent("CARD", otherCardSource),
		compono.WithGlobalComponent("TABLE", tableSource),
	),
)
```

There are two `CARD`s here and both work. A `CARD` call in the page resolves to the first one, and a `CARD` call inside `LAYOUT` resolves to the second one. `LAYOUT` cannot reach the page's `CARD`.

**Visibility:** A sub component can only be called from its owner's body and from its siblings' bodies. It is not visible from the converted source or from other global components. It neither overrides a same-named global component outside nor is overridden by one.

**Resolution:** A call inside a global component's body looks in this order:

1. The local components of the source the call is written in.
2. The global component's sub components.
3. The scope the global component is defined in, then outward through the scopes of its owners, and finally the global components given to `Convert` and registered with `RegisterGlobalComponent`.
4. Built-in components.

A global component without sub components resolves exactly as before: `Local > Global > Built-in`. Nesting depth is unlimited and the same rules apply at every level.

**Isolation:** `WithIsolatedScope()` stops step 3. Calls in the body of that global component and in the bodies of its sub components resolve only to locals, its sub components and built-ins. Giving it more than once has the same effect as giving it once.

**Syntactic resolution:** A call resolves in the scope of the source it is written in, not where it is rendered. If the page passes `{{ LAYOUT content = BODY }}`, the calls inside `BODY` resolve in the page's scope even though they are rendered inside `LAYOUT`. `LAYOUT`'s sub components are not visible from `BODY`.

**Built-in names:** Sub components override built-in components, and Compono does not check sub component names against the built-in list. Adding a new built-in in a later version does not change the output of a sub component with that name.

**Errors:** `Convert` returns an error when:

- a conversion option (`WithContext`, `WithErrorStylesheet`, `WithRendererHook`) is given to a global component at any depth, even with an empty value (`ErrConversionOptionInGlobal`),
- `WithIsolatedScope` is given directly to `Convert` (`ErrIsolatedScopeInConvert`),
- two sub components of the same owner share a name (`ErrDuplicateSubComponent`).

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

### IMAGE

Creates a semantic web image output from a `media` record and optional responsive variants.

Basic usage:

```
{{ IMAGE media = {
  url: "https://cdn.example.com/my-photo.jpg",
  width: 1600,
  height: 900,
  mime-type: "image/jpeg",
  variants: [
    {
      url: "https://cdn.example.com/my-photo-640.avif",
      width: 640,
      height: 360,
      mime-type: "image/avif"
    },
    {
      url: "https://cdn.example.com/my-photo-1280.avif",
      width: 1280,
      height: 720,
      mime-type: "image/avif"
    },
    {
      url: "https://cdn.example.com/my-photo-1600.avif",
      width: 1600,
      height: 900,
      mime-type: "image/avif"
    },
    {
      url: "https://cdn.example.com/my-photo-640.webp",
      width: 640,
      height: 360,
      mime-type: "image/webp"
    },
    {
      url: "https://cdn.example.com/my-photo-1280.webp",
      width: 1280,
      height: 720,
      mime-type: "image/webp"
    },
    {
      url: "https://cdn.example.com/my-photo-1600.webp",
      width: 1600,
      height: 900,
      mime-type: "image/webp"
    }
  ]
} alt = "My photo" }}
```

Output:
```html
<picture><source type="image/avif" srcset="https://cdn.example.com/my-photo-640.avif 640w, https://cdn.example.com/my-photo-1280.avif 1280w, https://cdn.example.com/my-photo-1600.avif 1600w"><source type="image/webp" srcset="https://cdn.example.com/my-photo-640.webp 640w, https://cdn.example.com/my-photo-1280.webp 1280w, https://cdn.example.com/my-photo-1600.webp 1600w"><img src="https://cdn.example.com/my-photo.jpg" alt="My photo" width="1600" height="900"></picture>
```

Without variants, `IMAGE` renders a plain `img` element:

```
{{ IMAGE media = {
  url: "https://cdn.example.com/avatar.png",
  width: 512,
  height: 512,
  mime-type: "image/png"
} alt = "Profile avatar" }}
```

Output:
```html
<img src="https://cdn.example.com/avatar.png" alt="Profile avatar" width="512" height="512">
```

`IMAGE` can be used inline or as a block component depending on where it is called:

```
Gallery cover: {{ IMAGE media = {
  url: "https://cdn.example.com/gallery-cover.jpg",
  width: 800,
  height: 450,
  mime-type: "image/jpeg"
} alt = "Gallery cover" }} is ready.
```

#### IMAGE Parameters

- `media` is required and must be a record with:
  - `url`
  - `width`
  - `height`
  - `mime-type`
  - optional `variants`
- `alt` is a string. Pass `alt=""` for decorative images.

Each item in `variants` must be a record with:

- `url`
- `width`
- `height`
- `mime-type`

Supported mime types:

- `image/jpeg`
- `image/png`
- `image/webp`
- `image/gif`
- `image/avif`

#### IMAGE Behavior

- `media` is always the fallback image source.
- HTML output is a `picture` element when variants are given, otherwise a plain `img` element.
- variants are grouped by first-seen `mime-type`, preserving the original group order.
- within each mime type group, `srcset` entries are sorted by ascending width.
- an empty `variants` array is valid and renders only the fallback `img`.
- all widths and heights must be greater than `0`.
- all variants must preserve the same aspect ratio as the main `media`.
- duplicate `mime-type` + `width` pairs are invalid.

When validation fails, Compono renders an error placeholder instead of silently producing invalid markup. Common IMAGE-specific errors include:

- `Invalid built-in arguments`
- `Unsupported mime-type`
- `Invalid dimension`
- `Duplicate variant`
- `Inconsistent aspect ratio`

### Removed Built-in Components

#### WEB_GRID

`WEB_GRID` was removed in v0.7. It described a grid layout through parameters such as columns, rows, areas and breakpoints. That is presentation, not meaning. Compono only carries semantic content, and layout belongs to the stylesheet of whoever renders the output. Keeping it would also have frozen its output DOM as part of a stable contract.

A `WEB_GRID` call now renders an `Unknown component` error. Local or global components named `WEB_GRID` are not affected and work as regular components.

#### NAVIGATION

`NAVIGATION` was removed in v0.7. It bundled a whole menu concept (`<nav>`, `<ul>`, `<li>` and `<a>`) into one built-in. It always produced an unordered list, so ordered navigations like breadcrumbs could not be expressed, and it could not carry attributes such as `aria-current` or `aria-label`. Its item records also used `label`/`target` instead of Compono's `text`/`url` naming.

A `NAVIGATION` call now renders an `Unknown component` error. Local or global components named `NAVIGATION` are not affected and work as regular components.

## Parameters

Components can accept parameters. Each parameter must have a **default value** defined in the component definition.

If a parameter value is not provided during the call, the **default value is used**. A parameter can also be marked as [required](#required-parameters).

```
{{ SAY_HELLO name="Jane" }}

~ SAY_HELLO name="John"
# Hello, {{ name }}!
```

### Supported Types

Supported parameter types:

- **String** → `name = "John"`
- **Integer** → `age = 25`
- **Bool** → `active = true`
- **Component** → `comp = COMP`
- **Array** → `items = ["Jane", 22, true, COMP]`
- **Record** → `config = { lang: "tr", for-admin: true }`

### Required Parameters

A parameter whose name is followed by `!` is required. A call that does not give it renders `Missing argument`:

```
{{ COVER media = context(media-by-alias/cover) }}

~ COVER media! = {} alt! = ""
{{ IMAGE media = media alt = alt }}
```

This renders `Missing argument` with the message `The parameter **alt** of component **COVER** is required.` When several are missing, the message is `The parameters **a**, **b** of component **[component]** are required.`

- The `!` is glued to the name: `alt! = ""`.
- The default value of a required parameter only declares its type. It is never used or resolved.
- An empty value satisfies the requirement. `alt = ""` is a given argument. The component decides what an empty value means.
- A required argument can be given in the call or bound with [Argument Binding](#argument-binding). Bound and given arguments are checked together.
- The error is shown on the call that misses the argument. If the called component comes from a component parameter, it is shown on the topmost call where the component value is written.
- `!` is for local and global components. Required parameters of built-in components are checked by their own errors.

### Parameter Definition Errors

A parameter definition must be `name = default` or `name! = default`. A definition without a default value (`~ X a`) or with any other shape (`~ X a ! = ""`, `~ X !a = ""`) renders `Invalid parameter definition` with the message `The parameter definition **[text]** of component **[component]** is invalid.` The error is shown on every call of the component, and the whole call is replaced by the error.

---

### Passing Parameters to Other Components

A parameter can be passed directly to another component call.

```
{{ USER age=31 }}

~ USER age=18
{{ ANOTHER_COMP another-integer-param=age }}

~ ANOTHER_COMP another-integer-param=0
Integer: *{{ another-integer-param }}*
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

### Argument Binding

When a component is passed as a value, it is not known who renders it and with which arguments. Arguments can be bound to a component value in parentheses:

```
{{ SITE_NAV menu-title = "Main Menu" }}

~ SITE_NAV menu-title = ""
{{ WRAPPER content = MENU(title = menu-title) }}

~ WRAPPER content = NO_MATTER
{{ content }}

~ MENU title = ""
## {{ title }}
```

Output:

```html
<h2>Main Menu</h2>
```

- The parentheses hold an argument list, just like a component call. Arguments are separated by spaces and can span multiple lines: `CARD(title = "Hello" tags = ["a", "b"])`.
- Values can be anything a call argument can be: literals, parameter references (including accessors such as `item.children`), `context(key)`, arrays, records and components.
- The name and the parenthesis must be adjacent and the parentheses cannot be empty. `CARD (title = "Hello")` and `CARD()` are not argument binding. Write only the name when no argument is bound.
- A bound value is resolved where it is written. In `MENU(title = menu-title)`, `menu-title` is the parameter of `SITE_NAV`, no matter where `MENU` is rendered.
- A bound component can be used anywhere a component value can: as a call argument, a default value, an array item or a record value. Bindings can be nested: `CARD(footer = FOOTER(text = "Bye"))`.
- Built-in components can be bound too: `content = LINK(text = text url = url)`.
- When a bound component is rendered, the bound arguments and the arguments given by the caller are used together. Parameters that are neither bound nor given use their default values.

```
{{ PAGE }}

~ PAGE card = CARD(title = "Bound title")
{{ card body = "Given by the caller" }}

~ CARD title = "" body = ""
## {{ title }}
{{ body }}
```

Output:

```html
<h2>Bound title</h2><p>Given by the caller</p>
```

#### Argument Binding Errors

- Binding a parameter that the component does not define renders `Unknown parameter`.
- Binding a value of the wrong type renders `Wrong argument type`.
- Required arguments of a bound component are validated with the bound and the given arguments together. This applies to built-in components and to [required parameters](#required-parameters) of local and global components.
- Giving an argument to a parameter that is already bound renders `Duplicate argument` with the message `The parameter **[name]** of component **[component]** is already bound.` There is no precedence rule between a bound and a given argument.

```
{{ WRAPPER content = CARD(title = "Bound") }}

~ WRAPPER content = NO_MATTER
{{ content title = "Given" }}

~ CARD title = ""
# {{ title }}
```

Errors are shown on the call the bound component value is written in. For a value bound in a default value, they are shown on the call that uses the default value.

---

### Parameter Visibility

A component only sees the parameters it defines. There is no parameter inheritance:

- a local component of a global component does not see the parameters of that global component
- a component does not see the parameters of the component calling it

Every value a component needs is passed to it as an argument. Referencing a parameter that is not defined by the component renders an `Unknown parameter` error.

```
{{ OUTER title = "Hello" }}

~ OUTER title = ""
{{ INNER }}

~ INNER
# {{ title }}
// Unknown parameter: INNER does not define title
```

Pass the value explicitly instead:

```
~ OUTER title = ""
{{ INNER title = title }}

~ INNER title = ""
# {{ title }}
```

The same applies to global components:

```
c.RegisterGlobalComponent("PROFILE_PAGE", []byte(`
name="Guest"

{{ PROFILE_CARD name = name }}

~ PROFILE_CARD name = ""
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

---

### Array Parameters
```
{{ WRAPPER names = ["John", "Jane"] }}

~ WRAPPER names = []
{{ SAY_HELLO name = names[0] }}
{{ SAY_HELLO name = names[1] }}

~ SAY_HELLO name = ""
# Hello **{{ name }}**!
```

Arrays do not have to be homogeneous.
```
~ COMP mix = ["Jane", 22, true, SAY_HELLO]
We can reach an element via index.
{{ mix[2] }}
// true
```

Arrays can be nested.
```
{{ TABLE data = [
  [1,2],
  [3,4],
]}}

~ TABLE data = []
{{ data[0][0] }} - {{ data[0][1] }}
{{ data[1][0] }} - {{ data[1][1] }}
```

---

### Record Parameters
Pass data as key - value

```
{{ COMP record = { title: "Hello", content: "Here Content" } }}

~ COMP record = {}
# {{ record.title }}
{{ record.content }}
```
Records can be nested
```
{{ COMP nested = {record: {key-1: "string", key-2: 123}, empty-record: {} } }}

~ COMP nested = {}
{{ nested.record.key-1 }} - {{ nested.record.key-2 }}
```

---

## Context

`context(key)` is a built-in reference mechanism for injecting immutable values at convert time.

Use it with `compono.WithContext`:

```go
type CurrentUser struct {
    FirstName string `compono:"first-name"`
    LastName  string `compono:"last-name"`
}

err := c.Convert(source, &buf, compono.WithContext(map[string]any{
    "app/version":   "1.2.0",
    "feature/live":  true,
    "stats/numbers": []int{10, 20, 30},
    "current-user": CurrentUser{
        FirstName: "Yunus",
        LastName:  "Emre",
    },
}))
```

Direct usage:

```
Version: {{ context(app/version) }}
```

It can also be used as:

- a component argument
- a default parameter value
- an array item
- a record value
- built-in component arguments

Example:

```
{{ LINK text=context(link/text) url=context(link/url) new-tab=context(link/new-tab) }}

~ GREETING name=context(current-user/first-name)
Hello **{{ name }}**!
```

If the resolved value is a record or array, you can keep using normal access syntax:

```
# {{ context(current-user).first-name }}
## {{ context(stats/numbers)[1] }}
```

### Context Keys

Keys are static and unquoted. They are made of `segment`s joined by `/`.

- segments may contain lowercase Latin letters, numbers, and `-`
- `/` cannot appear at the beginning or end
- repeated separators are invalid
- whitespace is allowed around the key inside `context(...)`

If `context()` is empty or the key format is invalid, it is treated as plain text instead of a context reference.

### Supported Go Types

`WithContext` supports:

- `string`
- `bool`
- `int`, `int8`, `int16`, `int32`, `int64`
- `[]T` and `[N]T`
- `map[string]T`
- structs

Notes:

- `nil` is not supported
- map keys must be `string`
- struct fields must be exported
- `compono` struct tags must be valid `kebab-case`
- if a struct field has no `compono` tag, its name is converted to `kebab-case`
- unsupported types such as floats, pointers, functions, and channels return a fatal conversion error
- fatal context injection errors are returned as `ErrUnsupportedType` or `ErrUnsupportedKeyNotation`

### Missing Keys and Errors

If a referenced key is not injected, Compono renders an error placeholder with:

- Title: `Unknown key`
- Message: `The key **[key]** is not injected.`

Error placement depends on how `context(...)` is used:

- direct usage always renders an inline error
- using it in a block component call renders a block error at the call site
- using it in an inline component call renders an inline error at the call site
- default values are resolved lazily, so no error is produced unless that parameter is actually used

## Renderer Hooks

Renderer hooks let an application inspect or replace the HTML output produced by the renderer for supported rendered units. They are registered per conversion with `compono.WithRendererHook`.

Hooks run after the renderer creates the default output for a markdown element or built-in component. The string returned by the hook becomes the output for that unit. When multiple hooks are registered, they run in registration order and each hook receives the output returned by the previous hook.

```go
import (
    "bytes"
    "strings"

    "github.com/umono-cms/compono"
    "github.com/umono-cms/compono/renderer/hook"
)

c := compono.New()

markdownHook := func(ctx hook.RendererHookContext) string {
    if ctx.Kind == hook.KindMarkdown && ctx.Name == "h1" {
        return strings.Replace(ctx.Output, "<h1>", `<h1 class="page-title">`, 1)
    }
    return ctx.Output
}

builtinHook := func(ctx hook.RendererHookContext) string {
    if ctx.Kind == hook.KindBuiltin && ctx.Name == "LINK" {
        if url, ok := ctx.Params.String("url"); ok && strings.HasPrefix(url, "https://example.com") {
            return strings.Replace(ctx.Output, "<a ", `<a data-internal="true" `, 1)
        }
    }
    return ctx.Output
}

var buf bytes.Buffer
err := c.Convert(
    []byte("# Hello\n\n{{ LINK text=\"Visit\" url=\"https://example.com\" }}"),
    &buf,
    compono.WithRendererHook(markdownHook),
    compono.WithRendererHook(builtinHook),
)
```

### Hook Context

Every hook receives a `hook.RendererHookContext`:

- `Kind`: where the output came from. `hook.KindMarkdown` is used for markdown elements, and `hook.KindBuiltin` is used for built-in components.
- `Name`: the rendered element or component name. Markdown names include `h1`, `h2`, `h3`, `h4`, `h5`, `h6`, `p`, `em`, `strong`, `link`, `inline-code`, and `code-block`. Built-in names use the component name, such as `LINK` or `IMAGE`.
- `Params`: raw resolved data for that rendered unit. Markdown elements expose values such as `content`, `text`, `url`, and `lang`. Built-in components expose successfully resolved arguments and defaults, including values coming from `context(key)`.
- `Output`: the current HTML output. Returning `ctx.Output` leaves it unchanged; returning another string replaces it.

`Params` is a `hook.Params` map of `hook.ParamValue`. Values can be read as strings, arrays, or records:

```go
imageHook := func(ctx hook.RendererHookContext) string {
    if ctx.Kind != hook.KindBuiltin || ctx.Name != "IMAGE" {
        return ctx.Output
    }

    media, ok := ctx.Params.Record("media")
    if !ok {
        return ctx.Output
    }

    mimeType, _ := media.String("mime-type")
    if mimeType == "image/avif" {
        return strings.Replace(ctx.Output, "<img ", `<img data-modern="true" `, 1)
    }
    return ctx.Output
}
```

Hook params are intentionally raw. For example, markdown link `text` and `url`, markdown `content`, and built-in arguments are provided before hook-level escaping or sanitizing. If a hook returns newly constructed HTML, that HTML is the responsibility of the application using the hook.

## Error Handling

Compono provides error feedback by rendering placeholders where errors occur.
Fatal errors during conversion stop the process and no output is produced.

### Error Elements

An error placeholder is rendered as a custom element that holds a closed [Declarative Shadow DOM](https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Elements/template#shadowrootmode). For example, `{{ FOO }}` on its own line, when `FOO` is not defined, renders (shown on multiple lines here, a single line in reality):

```html
<compono-error-block>
  <template shadowrootmode="closed">
    <link rel="stylesheet" href="/_umono/error.css">
    <div class="title">Unknown component</div>
    <div class="description">The component <strong>FOO</strong> is not defined or not registered.</div>
  </template>
</compono-error-block>
```

Inline errors use `compono-error-inline` with `span` elements instead of `div`. Block errors are rendered between blocks; inline errors are rendered where the error occurs, inside the surrounding paragraph or heading.

The text lives in a closed shadow root, so page stylesheets cannot select or style it. Styling comes only from the stylesheet linked inside the shadow root.

Title and description text is HTML-escaped. The `<strong>` emphasis is part of Compono's own structure.

### Error Stylesheet

`compono.WithErrorStylesheet` sets the stylesheet URL used by error elements:

```go
err := c.Convert(source, writer, compono.WithErrorStylesheet("/_umono/error.css"))
```

`/_umono/error.css` is only an example path. Compono does not serve or ship a stylesheet, so your application must serve one at the URL it passes.

- without the option, or with an empty string, no `<link>` is rendered
- the URL is not validated; it is HTML-escaped into `href`
- the option can be used once per conversion. A second `WithErrorStylesheet` makes `Convert` return a `*compono.ComponoError` with code `ErrErrorStylesheetAlreadySet`, even if one of the values is empty

## API Reference

### Core Methods

```go
// Create a new Compono instance
c := compono.New()

// Convert source to HTML
err := c.Convert(source []byte, writer io.Writer, opts ...compono.ConvertOption)

// Register a global component
err := c.RegisterGlobalComponent(name string, source []byte)

// Unregister a global component
err := c.UnregisterGlobalComponent(name string)

// Inject a global component for a single conversion
err := c.Convert(source, writer, compono.WithGlobalComponent(name, globalSource))

// Give a global component its own sub components and isolate its scope
err := c.Convert(source, writer, compono.WithGlobalComponent(name, globalSource,
    compono.WithIsolatedScope(),
    compono.WithGlobalComponent(subName, subSource),
))

// Inject convert-time context values
err := c.Convert(source, writer, compono.WithContext(map[string]any{
    "app/version": "1.2.0",
}))

// Register a renderer hook for a single conversion
err := c.Convert(source, writer, compono.WithRendererHook(func(ctx hook.RendererHookContext) string {
    return ctx.Output
}))

// Set the stylesheet URL linked inside error elements (once per conversion)
err := c.Convert(source, writer, compono.WithErrorStylesheet("/_umono/error.css"))
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

Inside a global component with sub components, the sub components come between the local components and the outer global components. See [Global Component Scopes](#global-component-scopes).

## License

MIT License - see [LICENSE](LICENSE) for details.
