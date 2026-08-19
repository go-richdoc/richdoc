# richdoc

A neutral, format-agnostic and widget-agnostic **document model** for rich
text, written in pure Go (CGO-free).

`richdoc` is the foundation of a multi-format rich-document system: a WYSIWYG
toolkit widget and converters for Markdown, LaTeX, ODT and RTF are all built on
top of this one model. It intentionally does no rendering, parsing or I/O — it
only defines the tree and a few small utilities to traverse, build, extract and
copy it.

## Model

A `Document` is an ordered slice of block nodes plus a format-agnostic
`map[string]string` of metadata. Blocks and inlines are **closed interface
sets**: each concrete type carries an unexported marker method, so consumers
(converters especially) can type-switch over them exhaustively.

- **Blocks**: `Heading`, `Paragraph`, `List` (`ListItem`), `CodeBlock`,
  `BlockQuote`, `Table` (`Cell`, `Alignment`), `ThematicBreak`, `MathBlock`,
  `RawBlock`.
- **Inlines**: `Text`, `Emph`, `Strong`, `Strikethrough`, `Code`, `Link`,
  `Image`, `Math`, `LineBreak`, `RawInline`.

`RawBlock`/`RawInline` carry format-specific passthrough text for round-trip
fidelity.

## Utilities

- `Walk(d *Document, v Visitor)` — depth-first traversal (`Enter`/`Leave`).
- `New() *Builder` — fluent construction with inline/structural constructor
  helpers.
- `PlainText(d *Document) string` — textual content for search and previews.
- `Clone(d *Document) *Document` — deep copy for undo snapshots and rewrites.

## Example

```go
doc := richdoc.New().
	Meta("title", "Hello").
	H(1, richdoc.Txt("Hello")).
	P(
		richdoc.Bold(richdoc.Txt("bold")),
		richdoc.Txt(" and "),
		richdoc.Italic(richdoc.Txt("italic")),
	).
	UList(true,
		richdoc.Item(richdoc.Paragraph{Inlines: []richdoc.Inline{richdoc.Txt("first")}}),
		richdoc.Item(richdoc.Paragraph{Inlines: []richdoc.Inline{richdoc.Txt("second")}}),
	).
	Doc()
```

## License

BSD-3-Clause. Copyright (c) the go-richdoc authors.
