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

### A cell can hold blocks (v0.5.0)

reST's grid tables allow full block content in a cell, HTML's `<td>` allows any
flow content, LaTeX's `tabular` allows a parbox. `Cell` therefore has both:

```go
type Cell struct {
	Inlines []Inline // the flattened view, for a consumer that predates Blocks
	Blocks  []Block  // the faithful content, when there is more than a paragraph
	ColSpan int
	RowSpan int
}
```

The duplication is deliberate and its contract is short: a producer that sets
`Blocks` must also set `Inlines`, a consumer reads whichever it can represent, and
`Cell.Content()` is that preference order in one call — `Blocks` when present,
otherwise the `Inlines` wrapped in a single `Paragraph`. A cell holding one
paragraph, which is nearly all of them, still sets `Inlines` alone.

It is measured, not speculative: over go-richdoc/rst's 1564-document corpus, 64
list items, 61 line-block lines, 35 literal blocks and 23 bullet lists in 24 files
were being flattened to a run of text. `Table.Caption []Inline` arrived with it,
for 24 captions in 13 files.

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
