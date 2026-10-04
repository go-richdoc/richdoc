// Copyright (c) the go-richdoc authors.
// SPDX-License-Identifier: BSD-3-Clause

package richdoc

// Clone returns a deep copy of d: the returned document shares no mutable
// state (slices or maps) with the original, so either may be mutated
// independently. It returns nil for a nil document.
//
// Editors use it for undo snapshots and converters for non-destructive
// rewrites.
func Clone(d *Document) *Document {
	if d == nil {
		return nil
	}
	out := &Document{Blocks: cloneBlocks(d.Blocks)}
	if d.Meta != nil {
		out.Meta = make(map[string]string, len(d.Meta))
		for k, v := range d.Meta {
			out.Meta[k] = v
		}
	}
	return out
}

func cloneBlocks(blocks []Block) []Block {
	if blocks == nil {
		return nil
	}
	out := make([]Block, len(blocks))
	for i, b := range blocks {
		out[i] = cloneBlock(b)
	}
	return out
}

func cloneBlock(b Block) Block {
	switch n := b.(type) {
	case Heading:
		n.Inlines = cloneInlines(n.Inlines)
		return n
	case Paragraph:
		n.Inlines = cloneInlines(n.Inlines)
		n.Classes = cloneStrings(n.Classes)
		return n
	case List:
		n.Items = cloneItems(n.Items)
		n.Classes = cloneStrings(n.Classes)
		return n
	case BlockQuote:
		n.Blocks = cloneBlocks(n.Blocks)
		n.Classes = cloneStrings(n.Classes)
		return n
	case Table:
		return cloneTable(n)
	case CodeBlock:
		// Text and Language are values; Classes is not, which is why this can
		// no longer fall to the default below.
		n.Classes = cloneStrings(n.Classes)
		return n
	default:
		// ThematicBreak, MathBlock, RawBlock hold no nested slices, so copying
		// the value is already a deep copy.
		return b
	}
}

// cloneStrings copies a Classes slice. Without it a clone SHARED the original's
// classes, so appending to one document's classes appended to the other's --
// exactly the aliasing Clone exists to prevent, and invisible until someone
// mutated a copy.
func cloneStrings(in []string) []string {
	if in == nil {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func cloneItems(items []ListItem) []ListItem {
	if items == nil {
		return nil
	}
	out := make([]ListItem, len(items))
	for i, it := range items {
		out[i] = ListItem{Blocks: cloneBlocks(it.Blocks)}
	}
	return out
}

func cloneTable(t Table) Table {
	out := Table{
		Caption: cloneInlines(t.Caption),
		Header:  cloneCells(t.Header),
		Classes: cloneStrings(t.Classes),
	}
	if t.Align != nil {
		out.Align = make([]Alignment, len(t.Align))
		copy(out.Align, t.Align)
	}
	if t.Rows != nil {
		out.Rows = make([][]Cell, len(t.Rows))
		for i, row := range t.Rows {
			out.Rows[i] = cloneCells(row)
		}
	}
	return out
}

func cloneCells(cells []Cell) []Cell {
	if cells == nil {
		return nil
	}
	out := make([]Cell, len(cells))
	for i, c := range cells {
		c.Inlines = cloneInlines(c.Inlines)
		c.Blocks = cloneBlocks(c.Blocks)
		out[i] = c
	}
	return out
}

func cloneInlines(inlines []Inline) []Inline {
	if inlines == nil {
		return nil
	}
	out := make([]Inline, len(inlines))
	for i, in := range inlines {
		out[i] = cloneInline(in)
	}
	return out
}

func cloneInline(in Inline) Inline {
	switch n := in.(type) {
	case Emph:
		n.Inlines = cloneInlines(n.Inlines)
		return n
	case Strong:
		n.Inlines = cloneInlines(n.Inlines)
		return n
	case Strikethrough:
		n.Inlines = cloneInlines(n.Inlines)
		return n
	case Link:
		n.Inlines = cloneInlines(n.Inlines)
		return n
	case Footnote:
		n.Blocks = cloneBlocks(n.Blocks)
		return n
	case Anchor:
		n.Inlines = cloneInlines(n.Inlines)
		return n
	case CrossRef:
		n.Inlines = cloneInlines(n.Inlines)
		return n
	case Code:
		n.Classes = cloneStrings(n.Classes)
		return n
	case Image:
		n.Classes = cloneStrings(n.Classes)
		return n
	default:
		// Text, Math, LineBreak and RawInline hold no nested slices, so copying
		// the value is already a deep copy.
		return in
	}
}
