// Copyright (c) the go-richdoc authors.
// SPDX-License-Identifier: BSD-3-Clause

package richdoc

// Visitor observes a depth-first traversal driven by [Walk].
//
// For every node Walk calls Enter before descending into that node's children
// and Leave once its subtree has been fully visited. Returning false from
// Enter skips the node's children (Leave is still called), which lets a
// visitor prune whole subtrees.
//
// Nodes are passed as any. The concrete dynamic types are the [Document]
// (passed as *Document), every [Block] and [Inline], and the structural
// [ListItem] and [Cell] containers, so a visitor can recover the full tree
// structure by type-switching.
type Visitor interface {
	Enter(node any) (descend bool)
	Leave(node any)
}

// Walk performs a depth-first traversal of d, driving v. It is a no-op when d
// is nil.
func Walk(d *Document, v Visitor) {
	if d == nil {
		return
	}
	if v.Enter(d) {
		for _, b := range d.Blocks {
			walkBlock(b, v)
		}
	}
	v.Leave(d)
}

func walkBlock(b Block, v Visitor) {
	if v.Enter(b) {
		switch n := b.(type) {
		case Heading:
			walkInlines(n.Inlines, v)
		case Paragraph:
			walkInlines(n.Inlines, v)
		case List:
			for _, it := range n.Items {
				walkItem(it, v)
			}
		case BlockQuote:
			for _, c := range n.Blocks {
				walkBlock(c, v)
			}
		case Table:
			for _, c := range n.Header {
				walkCell(c, v)
			}
			for _, row := range n.Rows {
				for _, c := range row {
					walkCell(c, v)
				}
			}
		}
	}
	v.Leave(b)
}

func walkItem(it ListItem, v Visitor) {
	if v.Enter(it) {
		for _, b := range it.Blocks {
			walkBlock(b, v)
		}
	}
	v.Leave(it)
}

func walkCell(c Cell, v Visitor) {
	if v.Enter(c) {
		walkInlines(c.Inlines, v)
	}
	v.Leave(c)
}

func walkInlines(inlines []Inline, v Visitor) {
	for _, in := range inlines {
		walkInline(in, v)
	}
}

func walkInline(in Inline, v Visitor) {
	if v.Enter(in) {
		switch n := in.(type) {
		case Emph:
			walkInlines(n.Inlines, v)
		case Strong:
			walkInlines(n.Inlines, v)
		case Strikethrough:
			walkInlines(n.Inlines, v)
		case Link:
			walkInlines(n.Inlines, v)
		case Footnote:
			for _, c := range n.Blocks {
				walkBlock(c, v)
			}
		case Anchor:
			walkInlines(n.Inlines, v)
		case CrossRef:
			walkInlines(n.Inlines, v)
		}
	}
	v.Leave(in)
}
