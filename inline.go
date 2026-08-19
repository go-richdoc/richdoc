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
