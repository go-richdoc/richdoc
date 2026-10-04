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
//
// Classes are the author's own presentational names for it; see [Classes].
type Paragraph struct {
	Inlines []Inline
	Classes []string
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
	Classes []string
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
	Classes  []string
}

// BlockQuote is a quotation containing nested blocks.
type BlockQuote struct {
	Blocks  []Block
	Classes []string
}

// Table is a simple grid with an optional header row.
//
// Align gives the per-column alignment; a shorter Align slice leaves the
// remaining columns at [AlignDefault]. Header may be empty for a headerless
// table. Rows is a list of rows, each a slice of cells.
//
// Caption is the table's own title — reST's ".. table:: Caption", HTML's
// <caption>, LaTeX's \caption — and is empty for a table that has none. It is
// inline content because every format that has a caption renders it as a line of
// text, not as a sequence of blocks.
type Table struct {
	Caption []Inline
	Align   []Alignment
	Header  []Cell
	Rows    [][]Cell
	Classes []string
}

// Cell is a single table cell.
//
// # Inlines and Blocks
//
// A cell can hold real block content: a list, a literal block, several
// paragraphs, even another table. Blocks carries it. Inlines carries the same
// content FLATTENED to a line of inlines, and remains what a consumer that does
// not know about Blocks reads.
//
// Both are filled for a cell whose content is more than one paragraph, and that
// duplication is the point: a converter written before Blocks existed keeps
// rendering the cell's words instead of silently rendering an empty cell. The
// contract is therefore
//
//   - a producer that sets Blocks MUST also set Inlines, to whatever degraded
//     rendering it would have produced before;
//   - a consumer reads Blocks when it can represent them and Inlines when it
//     cannot, which is what [Cell.Content] expresses in one call;
//   - the two must say the same thing. Nothing enforces it, so a producer that
//     edits one edits the other.
//
// A cell holding exactly one paragraph — the overwhelming common case — sets
// Inlines alone and leaves Blocks nil, so nothing about the simple case changes.
//
// # Spans
//
// ColSpan and RowSpan are the number of columns/rows this cell occupies.
// Zero, the default — what an existing Cell{Inlines: ...} literal or a
// [Td] call already produces, with no field for either — means the same
// as 1: an ordinary cell spanning nothing extra. A converter that has no
// notion of spanning cells at all (a plain CommonMark table, say) never
// needs to touch these fields to keep working correctly.
type Cell struct {
	Inlines []Inline
	Blocks  []Block
	ColSpan int
	RowSpan int
}

// Content returns the cell's content as blocks: Blocks when the producer filled
// it, and otherwise the Inlines wrapped in a single [Paragraph]. It returns nil
// for an empty cell.
//
// It exists so that a consumer able to render block content in a cell needs one
// call rather than a branch at every use, and so that the preference order is
// written down once.
func (c Cell) Content() []Block {
	if len(c.Blocks) > 0 {
		return c.Blocks
	}
	if len(c.Inlines) == 0 {
		return nil
	}
	return []Block{Paragraph{Inlines: c.Inlines}}
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
