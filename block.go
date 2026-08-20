// Copyright (c) the go-richdoc authors.
// SPDX-License-Identifier: BSD-3-Clause

package richdoc

// Block is a top-level or nested block-level node. The set of concrete block
// types is closed: only types defined in this package satisfy Block, which
// lets consumers type-switch exhaustively.
type Block interface {
	isBlock()
}

// Heading is a section heading. Level is 1..6, following the common HTML/
// Markdown convention (1 is the most prominent).
//
// ID is an optional anchor identifier for the heading (a Markdown heading
// anchor, a LaTeX \section immediately followed by \label). An empty ID means
// the heading carries no explicit anchor.
type Heading struct {
	Level   int
	ID      string
	Inlines []Inline
}

// Paragraph is a run of inline content forming a single logical paragraph.
type Paragraph struct {
	Inlines []Inline
}

// List is an ordered or unordered list.
//
// When Ordered is true, Start is the number of the first item (1 when unset).
// Tight indicates a list whose items should render without inter-item spacing,
// mirroring the CommonMark tight/loose distinction.
type List struct {
	Ordered bool
	Start   int
	Tight   bool
	Items   []ListItem
}

// ListItem is a single entry of a [List]. Items hold blocks, which makes
// arbitrary nesting (paragraphs, sub-lists, quotes, ...) possible.
type ListItem struct {
	Blocks []Block
}

// CodeBlock is a block of preformatted code. Language is an optional
// informational language tag (for example "go"); Text is the verbatim source
// including its internal newlines.
type CodeBlock struct {
	Language string
	Text     string
}

// BlockQuote is a quotation containing nested blocks.
type BlockQuote struct {
	Blocks []Block
}

// Table is a simple grid with an optional header row.
//
// Align gives the per-column alignment; a shorter Align slice leaves the
// remaining columns at [AlignDefault]. Header may be empty for a headerless
// table. Rows is a list of rows, each a slice of cells.
type Table struct {
	Align  []Alignment
	Header []Cell
	Rows   [][]Cell
}

// Cell is a single table cell holding inline content.
type Cell struct {
	Inlines []Inline
}

// ThematicBreak is a horizontal rule separating content.
type ThematicBreak struct{}

// MathBlock is display (block-level) mathematics, carrying its TeX source.
type MathBlock struct {
	TeX string
}

// RawBlock is a verbatim, format-specific block passthrough used to preserve
// round-trip fidelity for constructs the model does not represent natively.
// Format names the target format the Text belongs to (for example "latex" or
// "html"); a converter for a different format is free to drop it.
type RawBlock struct {
	Format string
	Text   string
}

// Alignment is the horizontal alignment of a table column.
type Alignment int

// Column alignments. AlignDefault leaves the choice to the renderer.
const (
	AlignDefault Alignment = iota
	AlignLeft
	AlignCenter
	AlignRight
)

func (Heading) isBlock()       {}
func (Paragraph) isBlock()     {}
func (List) isBlock()          {}
func (CodeBlock) isBlock()     {}
func (BlockQuote) isBlock()    {}
func (Table) isBlock()         {}
func (ThematicBreak) isBlock() {}
func (MathBlock) isBlock()     {}
func (RawBlock) isBlock()      {}
