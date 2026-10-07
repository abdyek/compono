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
    if _, err := c.Convert(source, &buf); err != nil {
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

### Comments

A line that starts with `//` is a comment line. Leading spaces and tabs are ignored. Comment lines are removed from the source, with their line breaks, before anything else is read, and they write nothing to the output:

```
# About
// TODO: confirm the founding year
Umono was founded in 2024.
```

Output:

```html
<h1>About</h1><p>Umono was founded in 2024.</p>
```

- A comment line is not an empty line, so it does not split a paragraph. `First line`, `// note`, `Second line` give `<p>First line<br>Second line</p>`.
- Comment lines are not lines of a component body. A component whose body is a comment line and one paragraph line is still inline.
- The content of a comment is not read. A call on a comment line is not called and produces no error, so `// {{ CARD }}` disables a line.
- There are no inline, multi-line or block comments. `Text // note`, `/* note */` and `<!-- note -->` are plain text.
- `//` does not start a comment inside a code block, inside a `{{ }}` unit that spans several lines, or after a heading's `# `.

Comments work in the converted source, in local components and in global components.

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

Global components are given to `Convert` with `WithGlobalComponent` and exist only in that conversion:

```go
c := compono.New()

c.Convert([]byte(`
# Page Title
Content here...
{{ FOOTER }}
`), &buf, compono.WithGlobalComponent("FOOTER", []byte(`© 2026 My Company`)))
```

A `Compono` value does not store global components. Pass them again to every conversion that uses them.

Global components can also have parameters:

```go
c.Convert(source, &buf, compono.WithGlobalComponent("BLOG_PAGE", []byte(`title="" content=""
## {{ title }}
{{ content }}`)))
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
3. The scope the global component is defined in, then outward through the scopes of its owners, and finally the global components given to `Convert` directly.
4. Built-in components.

A global component without sub components resolves exactly as before: `Local > Global > Built-in`. Nesting depth is unlimited and the same rules apply at every level.

**Isolation:** `WithIsolatedScope()` stops step 3. Calls in the body of that global component and in the bodies of its sub components resolve only to locals, its sub components and built-ins. Giving it more than once has the same effect as giving it once.

**Syntactic resolution:** A call resolves in the scope of the source it is written in, not where it is rendered. If the page passes `{{ LAYOUT content = BODY }}`, the calls inside `BODY` resolve in the page's scope even though they are rendered inside `LAYOUT`. `LAYOUT`'s sub components are not visible from `BODY`.

**Built-in names:** Sub components override built-in components, and Compono does not check sub component names against the built-in list. Adding a new built-in in a later version does not change the output of a sub component with that name.

**Errors:** `Convert` returns an error when:

- a conversion option (`WithContext`, `WithAttributeHook`) is given to a global component at any depth, even with an empty value (`ErrConversionOptionInGlobal`),
- `WithIsolatedScope` is given directly to `Convert` (`ErrIsolatedScopeInConvert`),
- two sub components of the same owner share a name (`ErrDuplicateSubComponent`),
- two global components given to `Convert` directly share a name (`ErrDuplicateGlobalComponent`).

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

When validation fails, the IMAGE call renders nothing instead of producing invalid markup, and `Convert` returns a diagnostic. Common IMAGE-specific diagnostic codes include:

- `invalid-built-in-arguments`
- `unsupported-mime-type`
- `invalid-dimension`
- `duplicate-variant`
- `inconsistent-aspect-ratio`

### Removed Built-in Components

#### WEB_GRID

`WEB_GRID` was removed in v0.7. It described a grid layout through parameters such as columns, rows, areas and breakpoints. That is presentation, not meaning. Compono only carries semantic content, and layout belongs to the stylesheet of whoever renders the output. Keeping it would also have frozen its output DOM as part of a stable contract.

A `WEB_GRID` call is now an `unknown-component` error and renders nothing. Local or global components named `WEB_GRID` are not affected and work as regular components.

#### NAVIGATION

`NAVIGATION` was removed in v0.7. It bundled a whole menu concept (`<nav>`, `<ul>`, `<li>` and `<a>`) into one built-in. It always produced an unordered list, so ordered navigations like breadcrumbs could not be expressed, and it could not carry attributes such as `aria-current` or `aria-label`. Its item records also used `label`/`target` instead of Compono's `text`/`url` naming.

A `NAVIGATION` call is now an `unknown-component` error and renders nothing. Local or global components named `NAVIGATION` are not affected and work as regular components.

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

A parameter whose name is followed by `!` is required. A call that does not give it is dropped and returns a `missing-argument` diagnostic:

```
{{ COVER media = context(media-by-alias/cover) }}

~ COVER media! = {} alt! = ""
{{ IMAGE media = media alt = alt }}
```

This drops the `COVER` call with the message `The parameter **alt** of component **COVER** is required.` When several are missing, the message is `The parameters **a**, **b** of component **[component]** are required.`

- The `!` is glued to the name: `alt! = ""`.
- The default value of a required parameter only declares its type. It is never used or resolved.
- An empty value satisfies the requirement. `alt = ""` is a given argument. The component decides what an empty value means.
- A required argument can be given in the call or bound with [Argument Binding](#argument-binding). Bound and given arguments are checked together.
- The call that misses the argument is dropped. If the called component comes from a component parameter, the call through the parameter (`{{ content }}`) is checked every time it is rendered and dropped where it is written.
- `!` is for local and global components. Required parameters of built-in components are checked by their own errors.

### Parameter Definition Errors

A parameter definition must be `name = default` or `name! = default`. A definition without a default value (`~ X a`) or with any other shape (`~ X a ! = ""`, `~ X !a = ""`) is an `invalid-parameter-definition` error with the message `The parameter definition **[text]** of component **[component]** is invalid.` Every call of the component is dropped and returns this diagnostic.

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

- Binding a parameter that the component does not define is an `unknown-parameter` error.
- Binding a value of the wrong type is a `wrong-argument-type` error.
- Required arguments of a bound component are validated with the bound and the given arguments together. This applies to built-in components and to [required parameters](#required-parameters) of local and global components.
- Giving an argument to a parameter that is already bound is a `duplicate-argument` error with the message `The parameter **[name]** of component **[component]** is already bound.` The call that gives the argument is dropped. There is no precedence rule between a bound and a given argument.

```
{{ WRAPPER content = CARD(title = "Bound") }}

~ WRAPPER content = NO_MATTER
{{ content title = "Given" }}

~ CARD title = ""
# {{ title }}
```

An error of a bound value drops the call the binding is written in. For a value bound in a default value, it drops the call that uses the default value. A missing or duplicate argument is found where the component value is called (`{{ content }}`) and drops that call.

---

### Parameter Visibility

A component only sees the parameters it defines. There is no parameter inheritance:

- a local component of a global component does not see the parameters of that global component
- a component does not see the parameters of the component calling it

Every value a component needs is passed to it as an argument. Referencing a parameter that is not defined by the component is an `unknown-parameter` error and drops the `{{ }}` unit.

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

```go
c.Convert(source, &buf, compono.WithGlobalComponent("PROFILE_PAGE", []byte(`
name="Guest"

{{ PROFILE_CARD name = name }}

~ PROFILE_CARD name = ""
## {{ name }}
Welcome to the profile page.
`)))
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

_, err := c.Convert(source, &buf, compono.WithContext(map[string]any{
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

If a referenced key is not injected, `Convert` returns a diagnostic with:

- Code: `unknown-key`
- Message: `The key **[key]** is not injected.`

What renders nothing depends on how `context(...)` is used:

- direct usage drops the `{{ context(...) }}` unit
- using it in an argument of a component call drops that call, also when the value reaches the argument through a parameter
- default values are resolved lazily, so no error is produced unless that parameter is actually used; then the `{{ }}` unit that uses it drops

## Attribute Hook

An attribute hook lets an application add HTML attributes to the output of built-in component calls, for example a theme's `class`. It is registered per conversion with `compono.WithAttributeHook`. The hook only returns attributes. Where they are written is defined by each built-in.

```go
_, err := c.Convert(pageSource, w,
	compono.WithGlobalComponent("LAYOUT", layoutSource,
		compono.WithIsolatedScope(),
		compono.WithGlobalComponent("MAIN_MENU", menuSource),
	),
	compono.WithAttributeHook(func(builtin string, chain []compono.Frame) map[string]string {
		if builtin == "LINK" && len(chain) > 0 && chain[len(chain)-1].Name == "MAIN_MENU" {
			return map[string]string{"class": "main-menu-link"}
		}
		return nil
	}),
)
```

For a `LINK` call in the body of `MAIN_MENU`, the hook receives `"LINK"` and this chain:

```go
[]compono.Frame{
	{Name: "LAYOUT", Kind: compono.FrameGlobal, ScopePath: []string{}},
	{Name: "MAIN_MENU", Kind: compono.FrameGlobal, ScopePath: []string{"LAYOUT"}},
}
```

### When the Hook Is Called

- Only for real built-in calls. A call that resolves to a local or global component is not a built-in call, even if its name is a built-in name. Markdown elements never call the hook.
- Once per call, before the built-in's output is written, in render order, on the goroutine running `Convert`.
- Never for a call that an error drops.

### The Chain

The chain is the call stack at the moment the built-in is rendered, ordered from the outermost frame to the innermost.

- Each frame has a `Name` and a `Kind`: `compono.FrameGlobal`, `compono.FrameLocal` or `compono.FrameBuiltin`.
- `ScopePath` is set only for global frames: the names of the global component's owners, outermost first (see [Global Component Scopes](#global-component-scopes)). It is an empty slice for a global in the root scope and `nil` for other kinds.
- The converted source is not a frame. A built-in written directly in the source gets an empty chain.
- The chain does not contain the rendered built-in itself. Its name is the `builtin` argument.
- The chain is dynamic. Content passed through a component parameter belongs to the frame it is rendered in, although its calls are resolved where they are written.
- The slices are only valid during the call. Copy them to keep them.

### Writing Attributes

- A `nil` or empty map writes nothing.
- Values are escaped and quoted. Attributes are written after Compono's own attributes, sorted by name, so the output is deterministic.
- Each attribute goes to the root element of the built-in, unless the built-in names a special point for it:
  - `LINK`: everything goes to `<a>`.
  - `IMAGE`: `sizes` goes to every element that carries `srcset` (each `<source>`), and the other attributes go to the root element. An `IMAGE` without variants has no `<source>`, so `sizes` is not written. This is not an error.
- Built-ins never write `class` or `sizes` themselves, so the hook's value is the only value. There is no merging.

### Hook Errors

`Convert` returns an error when:

- `WithAttributeHook` is given more than once in a conversion, even with `nil` (`ErrAttributeHookAlreadySet`). A `nil` hook given once is the same as no hook.
- it is given to a global component (`ErrConversionOptionInGlobal`),
- the hook returns a name that does not match `[a-z][a-z0-9-]*` (`ErrInvalidAttributeName`),
- the hook returns an attribute that Compono writes on that element itself, such as `href`, `target` or `rel` on `LINK`'s `<a>` (`ErrAttributeConflict`).

Nothing is written to the writer in these cases. The hook cannot return an error. Validating the values is the job of the application that gives the hook.

## Error Handling

There are two kinds of errors:

- **Fatal errors:** unexpected failures and invalid options. `Convert` returns them as `error` and writes no output.
- **Diagnostics:** syntax and logic errors in the source. An error drops the part of the output it belongs to: that part renders nothing. The rest of the output is written and valid, and `Convert` returns a `compono.Diagnostic` for the dropped part.

Compono never writes an error into the output. Who sees an error, and how, is up to the application.

```go
diagnostics, err := c.Convert(source, &buf)
if err != nil {
    return err
}
for _, d := range diagnostics {
    fmt.Printf("%v %d:%d %s: %s\n", d.Source, d.Range.Start.Line, d.Range.Start.Column, d.Code, d.Message)
}
```

### Diagnostics

A diagnostic has these fields:

- `Code`: the kind of the error, the kebab-case form of its title, such as `unknown-component`. Codes are exported as constants (`compono.CodeUnknownComponent`) and are never renamed. Several situations may share a code; the message tells them apart. Programs should look at the code, not the message.
- `Message`: an English description. Values are emphasized with `**`. It is not HTML; escaping it is the job of whoever shows it.
- `Source`: the source the dropped part is written in. It is empty for the converted source. For a global component it is its scope path ending with its own name: `[LAYOUT CARD]` for the sub component `CARD` of `LAYOUT`.
- `Range`: the `[Start, End)` range of the dropped part in `Source`. Each end has a 0-based byte `Offset`, a 1-based `Line` and a 1-based `Column` counted in runes. Positions are relative to the source as given: comment lines, a global's parameter line and leading blank lines are counted.
- `Calls`: the component calls around the dropped part, outermost first, each with its `Name`, `Kind` (`builtin`, `global` or `local`), `Source` and `Range`. It is empty when the dropped part is in the converted source itself.

### What Is Dropped

An error drops the smallest part it belongs to:

- **A component call** (built-in, global or local) for its own errors: an unknown component, argument errors, built-in specific errors, an invalid parameter definition of the called component and an infinite call. An error of an argument's value also drops the call the argument belongs to.
- **A `{{ }}` unit** (a parameter reference or `context(key)`) for its own errors.
- **A markdown link** for an error in its address. An error of a `{{ }}` unit in the link text drops only that unit: `[Docs {{ x }}](/docs)` renders `<a href="/docs">Docs </a>` when `x` cannot be used.

The paragraph, heading or component body around a dropped part is still rendered. For example, `Hello {{ FOO }} world` with an undefined `FOO` renders `<p>Hello  world</p>`.

An error that depends on the value of a parameter is found where the value is used, separately for every render. A `{{ }}` unit with an index out of range, an unknown record key, an array or record used directly, or a block component used inline drops only itself, in the component where it is written. A call whose argument comes from a parameter, such as an `IMAGE` whose `media` is passed down or a call that forwards a value of the wrong type, is checked the same way and drops where it is written. The same unit can drop in one call and render in another:

```
{{ ITEM list = [1] }}

{{ ITEM list = [] }}

~ ITEM list = []
Item: {{ list[0] }}
```

renders `<p>Item: 1</p><p>Item: </p>` and returns one `array-index-out-of-range` diagnostic whose `Calls` holds the second `ITEM` call.

A call that enters a component already being rendered with the same component values never ends. It is an `infinite-component-call` and drops where it is written, so in a cycle only the call that closes it drops. A component that calls itself with other component values is not a loop.

### Rules

- `Convert` returns one diagnostic for each dropped part. A global component called 10 times with an error inside returns 10 diagnostics, each with its own `Calls`; grouping them is up to the application.
- A part with several errors returns one diagnostic.
- An error that drops nothing returns no diagnostic, such as one inside a component that is never called.
- Diagnostics are returned in the order of the dropped parts in the output. The same source and options always return the same diagnostics in the same order.
- A diagnostic is not a failure: the output is written and valid. `error` is only for fatal errors, and then no diagnostics are returned.

## API Reference

### Core Methods

```go
// Create a new Compono instance
c := compono.New()

// Convert source to HTML and get the diagnostics of the conversion
diagnostics, err := c.Convert(source []byte, writer io.Writer, opts ...compono.ConvertOption)

// Give a global component for a single conversion
_, err := c.Convert(source, writer, compono.WithGlobalComponent(name, globalSource))

// Give a global component its own sub components and isolate its scope
_, err := c.Convert(source, writer, compono.WithGlobalComponent(name, globalSource,
    compono.WithIsolatedScope(),
    compono.WithGlobalComponent(subName, subSource),
))

// Inject convert-time context values
_, err := c.Convert(source, writer, compono.WithContext(map[string]any{
    "app/version": "1.2.0",
}))

// Add attributes to built-in component calls (once per conversion)
_, err := c.Convert(source, writer, compono.WithAttributeHook(func(builtin string, chain []compono.Frame) map[string]string {
    return nil
}))
```

### Concurrency

A `Compono` value is stateless. `Convert` depends only on the source and the options, and it changes nothing in the `Compono` value. Everything that belongs to a conversion (global components, context, attribute hook) is given to `Convert` as an option and applies only to that conversion. One `Compono` value can be shared and its `Convert` called from several goroutines at once, with different options. The configuration does not change after `New`.

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
c.Convert(source, &buf, compono.WithGlobalComponent("LINK", []byte(`Custom link behavior`)))
```

Now all `{{ LINK }}` calls in that conversion will use your global definition instead of the built-in one.

This allows you to customize or extend built-in components without modifying the library.

Inside a global component with sub components, the sub components come between the local components and the outer global components. See [Global Component Scopes](#global-component-scopes).

## License

MIT License - see [LICENSE](LICENSE) for details.
