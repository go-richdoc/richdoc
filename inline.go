// Copyright (c) the go-richdoc authors.
// SPDX-License-Identifier: BSD-3-Clause

package richdoc

// Inline is an inline-level node. Like [Block], the set of concrete inline
// types is closed, so consumers can type-switch exhaustively.
type Inline interface {
	isInline()
}

// Text is a run of literal text.
type Text struct {
	Value string
}

// Emph is emphasized (conventionally italic) inline content.
type Emph struct {
	Inlines []Inline
}

// Strong is strongly emphasized (conventionally bold) inline content.
type Strong struct {
	Inlines []Inline
}

// Strikethrough is struck-out inline content.
type Strikethrough struct {
	Inlines []Inline
}

// Code is an inline code span. Value is the verbatim code.
type Code struct {
	Value string
}

// Link is a hyperlink wrapping inline content. Title is an optional advisory
// title (for example a tooltip).
type Link struct {
	URL     string
	Title   string
	Inlines []Inline
}

// Image is an inline image reference. Alt is the textual alternative and Title
// an optional advisory title.
type Image struct {
	URL   string
	Alt   string
	Title string
}

// Math is inline mathematics, carrying its TeX source.
type Math struct {
	TeX string
}

// LineBreak is a hard line break within a block.
type LineBreak struct{}

// RawInline is a verbatim, format-specific inline passthrough used to preserve
// round-trip fidelity, analogous to [RawBlock]. Format names the target format
// the Text belongs to.
type RawInline struct {
	Format string
	Text   string
}

// Footnote is a footnote placed inline, whose content is block-level (LaTeX
// \footnote, an ODF footnote, a Markdown [^id] reference with its definition).
// Blocks holds the note body; it appears at the position the note is
// referenced, and a writer is free to relocate the body to the page or
// document end.
type Footnote struct {
	Blocks []Block
}

// Anchor is a labeled target attached to inline content: the destination a
// [CrossRef] points at (a LaTeX \label, an ODF bookmark, a Markdown heading
// anchor target). ID is the label; Inlines is the content the label marks and
// may be empty for a point target that carries no visible text of its own.
type Anchor struct {
	ID      string
	Inlines []Inline
}

// RefKind distinguishes the two kinds of reference a [CrossRef] can be.
type RefKind int

// Reference kinds. RefLabel is a cross-reference to a labeled target (LaTeX
// \ref/\eqref, a Markdown link to an internal id); RefCite is a citation of a
// bibliographic key (LaTeX \cite).
const (
	RefLabel RefKind = iota
	RefCite
)

// CrossRef is a reference to an [Anchor]/label or a bibliographic citation.
// Target is the label or citation key it resolves to and Kind selects between
// the two. Inlines is the visible text; when it is empty the renderer or
// writer supplies the resolved number or label.
type CrossRef struct {
	Target  string
	Kind    RefKind
	Inlines []Inline
}

func (Text) isInline()          {}
func (Emph) isInline()          {}
func (Strong) isInline()        {}
func (Strikethrough) isInline() {}
func (Code) isInline()          {}
func (Link) isInline()          {}
func (Image) isInline()         {}
func (Math) isInline()          {}
func (LineBreak) isInline()     {}
func (RawInline) isInline()     {}
func (Footnote) isInline()      {}
func (Anchor) isInline()        {}
func (CrossRef) isInline()      {}
