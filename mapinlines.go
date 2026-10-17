// Copyright (c) the go-richdoc authors.
// SPDX-License-Identifier: BSD-3-Clause

package richdoc

// MapInlines returns a copy of blocks with every inline node passed through f,
// in document order, however deeply it is nested: inside a paragraph, a
// heading, a list item, a table cell, a quotation, a link, an emphasis.
//
// It is the rewriting counterpart of [Walk]. Walk OBSERVES — nodes are passed
// as values, so a visitor that edits one edits a copy — and that is the right
// shape for counting and for collecting. Replacing a node needs the tree
// rebuilt around it, which is what this does.
//
// Two converters asked for it before it existed, both with the same job:
// [github.com/go-richdoc/epub] resolves an image's href against the ZIP entry
// it names, and [github.com/go-richdoc/email] resolves a cid: reference
// against the MIME part it names. Each had written the same thirty lines of
// type switch, and each had the same hazard — a container the switch forgot
// keeps the un-rewritten node, silently, because nothing is missing from the
// output. One copy, in the package that owns the node set, cannot forget a
// container that the node set gains.
//
// f sees every inline, including those inside the inline containers it may
// itself return; it is called exactly once per node of the input. Returning
// the node unchanged is free.
//
// Blocks with no inline content — a code block, a thematic break, raw content,
// a maths block — come through as they are.
func MapInlines(blocks []Block, f func(Inline) Inline) []Block {
	if blocks == nil {
		return nil
	}
	out := make([]Block, len(blocks))
	for i, b := range blocks {
		out[i] = mapBlockInlines(b, f)
	}
	return out
}

// MapDocumentInlines is [MapInlines] over a whole document, metadata and all
// (the metadata is copied, not mapped: it is strings and not nodes). It
// returns nil for a nil document, as [Clone] does.
func MapDocumentInlines(d *Document, f func(Inline) Inline) *Document {
	if d == nil {
		return nil
	}
	out := &Document{Blocks: MapInlines(d.Blocks, f)}
	if d.Meta != nil {
		out.Meta = make(map[string]string, len(d.Meta))
		for k, v := range d.Meta {
			out.Meta[k] = v
		}
	}
	return out
}

func mapBlockInlines(b Block, f func(Inline) Inline) Block {
	switch n := b.(type) {
	case Heading:
		n.Inlines = mapInlineSlice(n.Inlines, f)
		return n
	case Paragraph:
		n.Inlines = mapInlineSlice(n.Inlines, f)
		return n
	case BlockQuote:
		n.Blocks = MapInlines(n.Blocks, f)
		return n
	case List:
		if n.Items != nil {
			items := make([]ListItem, len(n.Items))
			for i, it := range n.Items {
				it.Blocks = MapInlines(it.Blocks, f)
				items[i] = it
			}
			n.Items = items
		}
		return n
	case Table:
		n.Caption = mapInlineSlice(n.Caption, f)
		n.Header = mapCellInlines(n.Header, f)
		if n.Rows != nil {
			rows := make([][]Cell, len(n.Rows))
			for i, r := range n.Rows {
				rows[i] = mapCellInlines(r, f)
			}
			n.Rows = rows
		}
		return n
	}
	// CodeBlock, ThematicBreak, MathBlock, RawBlock: nothing inline inside.
	return b
}

func mapCellInlines(cells []Cell, f func(Inline) Inline) []Cell {
	if cells == nil {
		return nil
	}
	out := make([]Cell, len(cells))
	for i, c := range cells {
		// ⛔ BOTH fields, because Cell's contract is that they say the same
		// thing: a rewrite that touched only Inlines would leave a consumer
		// reading Blocks with the old node, and vice versa.
		c.Inlines = mapInlineSlice(c.Inlines, f)
		c.Blocks = MapInlines(c.Blocks, f)
		out[i] = c
	}
	return out
}

func mapInlineSlice(in []Inline, f func(Inline) Inline) []Inline {
	if in == nil {
		return nil
	}
	out := make([]Inline, len(in))
	for i, node := range in {
		out[i] = f(mapInlineChildren(node, f))
	}
	return out
}

// mapInlineChildren rewrites what is INSIDE an inline container, before the
// container itself is handed to f: f therefore sees a node whose children are
// already rewritten, which is what lets it replace a link by its contents or
// an emphasis by a plain run without the children being visited twice.
func mapInlineChildren(node Inline, f func(Inline) Inline) Inline {
	switch n := node.(type) {
	case Emph:
		n.Inlines = mapInlineSlice(n.Inlines, f)
		return n
	case Strong:
		n.Inlines = mapInlineSlice(n.Inlines, f)
		return n
	case Strikethrough:
		n.Inlines = mapInlineSlice(n.Inlines, f)
		return n
	case Link:
		n.Inlines = mapInlineSlice(n.Inlines, f)
		return n
	case Anchor:
		n.Inlines = mapInlineSlice(n.Inlines, f)
		return n
	case CrossRef:
		n.Inlines = mapInlineSlice(n.Inlines, f)
		return n
	case Footnote:
		// ⛔ A footnote is an INLINE whose content is blocks — it sits where it
		// is referenced and carries paragraphs — so the recursion crosses back
		// into MapInlines here. A rewrite that stopped at it would leave
		// everything in every note untouched, which is exactly the silent half
		// of this hazard.
		n.Blocks = MapInlines(n.Blocks, f)
		return n
	}
	// Text, Code, Image, Math, LineBreak, RawInline: leaves.
	return node
}
