// Copyright (c) the go-richdoc authors.
// SPDX-License-Identifier: BSD-3-Clause

package richdoc

import "strings"

// PlainText returns the textual content of d with block-level nodes separated
// by newlines. It is intended for search, previews and tests, not for
// faithful rendering.
//
// It concatenates the values of [Text] and [Code] inlines and of [CodeBlock]
// blocks, descending through every container (headings, lists, quotes, table
// cells, and emphasis-like inlines). A [Footnote] contributes its body text
// inline at the position it occurs, because footnotes are document text a
// search should find; an [Anchor] and a [CrossRef] contribute their visible
// inlines but not their identifiers. Nodes that carry no literal text in that
// sense contribute nothing: [ThematicBreak], [LineBreak], [Image], [Math],
// [MathBlock], [RawInline] and [RawBlock]. It returns "" for a nil document.
func PlainText(d *Document) string {
	if d == nil {
		return ""
	}
	return blocksText(d.Blocks)
}

func blocksText(blocks []Block) string {
	parts := make([]string, 0, len(blocks))
	for _, b := range blocks {
		if s := blockText(b); s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, "\n")
}

func blockText(b Block) string {
	switch n := b.(type) {
	case Heading:
		return inlinesText(n.Inlines)
	case Paragraph:
		return inlinesText(n.Inlines)
	case CodeBlock:
		return n.Text
	case BlockQuote:
		return blocksText(n.Blocks)
	case List:
		parts := make([]string, 0, len(n.Items))
		for _, it := range n.Items {
			if s := blocksText(it.Blocks); s != "" {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, "\n")
	case Table:
		return tableText(n)
	}
	return ""
}

func tableText(t Table) string {
	var parts []string
	add := func(cells []Cell) {
		for _, c := range cells {
			if s := inlinesText(c.Inlines); s != "" {
				parts = append(parts, s)
			}
		}
	}
	add(t.Header)
	for _, row := range t.Rows {
		add(row)
	}
	return strings.Join(parts, " ")
}

func inlinesText(inlines []Inline) string {
	var sb strings.Builder
	for _, in := range inlines {
		sb.WriteString(inlineText(in))
	}
	return sb.String()
}

func inlineText(in Inline) string {
	switch n := in.(type) {
	case Text:
		return n.Value
	case Code:
		return n.Value
	case Emph:
		return inlinesText(n.Inlines)
	case Strong:
		return inlinesText(n.Inlines)
	case Strikethrough:
		return inlinesText(n.Inlines)
	case Link:
		return inlinesText(n.Inlines)
	case Footnote:
		return blocksText(n.Blocks)
	case Anchor:
		return inlinesText(n.Inlines)
	case CrossRef:
		return inlinesText(n.Inlines)
	}
	return ""
}
