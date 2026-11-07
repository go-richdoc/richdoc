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
list items, 47 lines in 14 line blocks, 35 literal blocks and 23 bullet lists in 24 files
were being flattened to a run of text. `Table.Caption []Inline` arrived with it,
for 24 captions in 13 files.

## Utilities

- `Walk(d *Document, v Visitor)` — depth-first traversal (`Enter`/`Leave`).
- `New() *Builder` — fluent construction with inline/structural constructor
  helpers.
- `PlainText(d *Document) string` — textual content for search and previews.
- `Clone(d *Document) *Document` — deep copy for undo snapshots and rewrites.

## `bombtest` — the gate the readers' bomb tests run through

A bomb test feeds a reader something enormous and asserts it is **refused**.
Five readers in this family needed the same three guards, each had its own copy,
and the copies are how the mistakes got made. So the questions live in
[`richdoc/bombtest`](https://pkg.go.dev/github.com/go-richdoc/richdoc/bombtest)
as fields with **no defaults**, and `Run` refuses a `Bound` that left any of
them unanswered *before* the bomb allocates:

```go
bombtest.Run(t, bombtest.Bound{
    Why:        "MaxSourceElements refuses before x/net/html builds anything",
    AsksAtMost: 1 << 20,   // the figure that NAMES a regression
    HeldAtMost: 256 << 20, // the host's safety net
}, func() error { _, err := html.Parse(src, lim); return err })
```

⛔ **Two bounds, because a held-only bound protects by avoidance and hides what
it avoids.** `GOMEMLIMIT` is a GC target, not an allocation refusal: a reader
whose ceiling regressed does not reserve more memory under it, it makes the
collector work harder — so the failure arrives as a test that takes minutes, and
a slow test reads as a slow machine. `AsksAtMost` reads `TotalAlloc`, which is
cumulative and cannot be collected away, so it fires at once with both numbers.

⛔ **A delta of `Sys` is not expressible**, which is the mistake that prompted
the package: `Sys` never shrinks, so `after - before` reports what a call *added*
to whatever grew the arena before it. The same bomb measured **94 736 400** bytes
run first and **40 386 560** run second, and a self-check for exactly this read
164 292 904 − 164 292 904 = **0**, passing alone and failing last. `Result.Held`
is absolute; `Result.Asked` is the `TotalAlloc` difference.

⛔ The refusal assertion **cannot be switched off** — a bomb test that passes on
an accepted document measures nothing. `Unbounded` is the derogation for one
enormous allocation, which a soft limit cannot hold anyway; it drops exactly one
check, **stamps** every run with `⛔ UNBOUNDED` and the reason, and leaves the
rest in force.

⛔ And `ulimit -v`, the hard version of `HeldAtMost`, is a **no-op on darwin**
(RLIMIT_AS is ignored), so a shell guard protects CI and nothing on the machine
where a bomb gets written. Hence the Go-side limit, and hence `Result.Held`
being checked rather than trusted.

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

## Toolchain

Requires **Go 1.27.1**, which is the version the CI workflow pins. The two were out of
step — the module asked for 1.26.4 while CI already ran 1.27.1 — and that difference
decides two things a reader of this repository should not have to guess: 1.27 counts
statements more finely, so a coverage figure from an older toolchain is an upper bound
rather than a measurement, and its `gofmt` reindents a composite literal inside a
multi-value return, which an older local `gofmt` reports as clean. Both were checked
with 1.27.1 here: coverage is unchanged at 100%, and the tree needs no reformatting.

## License

BSD-3-Clause. Copyright (c) the go-richdoc authors.
