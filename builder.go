// Copyright (c) the go-richdoc authors.
// SPDX-License-Identifier: BSD-3-Clause

package richdoc

// Builder incrementally assembles a [Document] with a fluent, chainable API.
// Every method appends a top-level block and returns the receiver, and [New]
// starts an empty builder:
//
//	doc := richdoc.New().
//		H(1, richdoc.Txt("Title")).
//		P(richdoc.Bold(richdoc.Txt("bold")), richdoc.Txt(" and "), richdoc.Italic(richdoc.Txt("italic"))).
//		Doc()
//
// Inline and structural values are produced by the constructor helpers in this
// file ([Txt], [Bold], [Italic], [Item], [Td], ...); the plain struct literals
// remain available for anything the helpers do not cover.
//
// A Builder is not safe for concurrent use.
type Builder struct {
	blocks []Block
	meta   map[string]string
}

// New returns an empty [Builder].
func New() *Builder { return &Builder{} }

// Doc finalizes construction and returns the assembled document.
func (b *Builder) Doc() *Document {
	d := &Document{Blocks: b.blocks}
	if b.meta != nil {
		d.Meta = b.meta
	}
	return d
}

// Meta sets a metadata key/value pair (for example "title" or "author").
func (b *Builder) Meta(key, value string) *Builder {
	if b.meta == nil {
		b.meta = make(map[string]string)
	}
	b.meta[key] = value
	return b
}

// Add appends arbitrary pre-built blocks, an escape hatch for constructs the
// typed methods below do not cover.
func (b *Builder) Add(blocks ...Block) *Builder {
	b.blocks = append(b.blocks, blocks...)
	return b
}

// H appends a [Heading] of the given level (1..6).
func (b *Builder) H(level int, inlines ...Inline) *Builder {
	b.blocks = append(b.blocks, Heading{Level: level, Inlines: inlines})
	return b
}

// P appends a [Paragraph].
func (b *Builder) P(inlines ...Inline) *Builder {
	b.blocks = append(b.blocks, Paragraph{Inlines: inlines})
	return b
}

// CodeBlock appends a [CodeBlock] with an optional language tag.
func (b *Builder) CodeBlock(language, text string) *Builder {
	b.blocks = append(b.blocks, CodeBlock{Language: language, Text: text})
	return b
}

// Quote appends a [BlockQuote] wrapping the given blocks.
func (b *Builder) Quote(blocks ...Block) *Builder {
	b.blocks = append(b.blocks, BlockQuote{Blocks: blocks})
	return b
}

// UList appends an unordered [List].
func (b *Builder) UList(tight bool, items ...ListItem) *Builder {
	b.blocks = append(b.blocks, List{Start: 1, Tight: tight, Items: items})
	return b
}

// OList appends an ordered [List] starting at start (clamped to a minimum of
// 1).
func (b *Builder) OList(start int, tight bool, items ...ListItem) *Builder {
	if start < 1 {
		start = 1
	}
	b.blocks = append(b.blocks, List{Ordered: true, Start: start, Tight: tight, Items: items})
	return b
}

// Table appends a [Table] with the given column alignments, header cells and
// rows. Any argument may be nil for an unaligned, headerless or empty table.
func (b *Builder) Table(align []Alignment, header []Cell, rows [][]Cell) *Builder {
	b.blocks = append(b.blocks, Table{Align: align, Header: header, Rows: rows})
	return b
}

// HR appends a [ThematicBreak].
func (b *Builder) HR() *Builder {
	b.blocks = append(b.blocks, ThematicBreak{})
	return b
}

// MathBlock appends a display-math [MathBlock] carrying TeX source.
func (b *Builder) MathBlock(tex string) *Builder {
	b.blocks = append(b.blocks, MathBlock{TeX: tex})
	return b
}

// RawBlock appends a [RawBlock] passthrough for the named format.
func (b *Builder) RawBlock(format, text string) *Builder {
	b.blocks = append(b.blocks, RawBlock{Format: format, Text: text})
	return b
}

// Inline constructor helpers.
//
// These have short names distinct from the node types they build, because a
// package-level function cannot share a name with a type. Each documents the
// type it returns.

// Txt builds a [Text] inline.
func Txt(value string) Text { return Text{Value: value} }

// Bold builds a [Strong] (bold) inline.
func Bold(inlines ...Inline) Strong { return Strong{Inlines: inlines} }

// Italic builds an [Emph] (italic) inline.
func Italic(inlines ...Inline) Emph { return Emph{Inlines: inlines} }

// Strike builds a [Strikethrough] inline.
func Strike(inlines ...Inline) Strikethrough { return Strikethrough{Inlines: inlines} }

// Mono builds an inline [Code] span.
func Mono(value string) Code { return Code{Value: value} }

// Href builds a [Link] inline.
func Href(url, title string, inlines ...Inline) Link {
	return Link{URL: url, Title: title, Inlines: inlines}
}

// Img builds an [Image] inline.
func Img(url, alt, title string) Image { return Image{URL: url, Alt: alt, Title: title} }

// InlineMath builds a [Math] inline carrying TeX source.
func InlineMath(tex string) Math { return Math{TeX: tex} }

// Br builds a hard [LineBreak] inline.
func Br() LineBreak { return LineBreak{} }

// RawI builds a [RawInline] passthrough for the named format.
func RawI(format, text string) RawInline { return RawInline{Format: format, Text: text} }

// Structural constructor helpers.

// Item builds a [ListItem] holding the given blocks.
func Item(blocks ...Block) ListItem { return ListItem{Blocks: blocks} }

// Td builds a table [Cell] holding the given inlines.
func Td(inlines ...Inline) Cell { return Cell{Inlines: inlines} }
