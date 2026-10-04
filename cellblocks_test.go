// Copyright (c) the go-richdoc authors.
// SPDX-License-Identifier: BSD-3-Clause

package richdoc

import (
	"reflect"
	"testing"
)

// blockCell is the shape this version exists for: a cell holding real blocks AND
// the flattened inline view a consumer that predates Blocks still reads.
func blockCell() Cell {
	return Cell{
		Inlines: []Inline{Txt("one"), Txt("\n\n"), Txt("two")},
		Blocks: []Block{
			Paragraph{Inlines: []Inline{Txt("one")}},
			CodeBlock{Language: "go", Text: "two"},
		},
	}
}

// TestCellContentPrefersBlocks pins the preference order, which is the whole of
// the contract a consumer has to know: Blocks when the producer filled it, the
// Inlines wrapped in one Paragraph otherwise, nil for an empty cell.
func TestCellContentPrefersBlocks(t *testing.T) {
	if got := blockCell().Content(); !reflect.DeepEqual(got, blockCell().Blocks) {
		t.Errorf("Content() = %#v, want the cell's Blocks", got)
	}
	want := []Block{Paragraph{Inlines: []Inline{Txt("just text")}}}
	if got := Td(Txt("just text")).Content(); !reflect.DeepEqual(got, want) {
		t.Errorf("Content() = %#v, want %#v", got, want)
	}
	if got := (Cell{}).Content(); got != nil {
		t.Errorf("Content() of an empty cell = %#v, want nil", got)
	}
}

// TestCloneDoesNotShareCellBlocksOrCaption is the aliasing test every container
// field in this package gets: Clone is a DEEP copy, so mutating the copy in place
// must not reach the original. Without the two clone lines this fails on both.
func TestCloneDoesNotShareCellBlocksOrCaption(t *testing.T) {
	orig := New().Add(Table{
		Caption: []Inline{Txt("Caption")},
		Header:  []Cell{Td(Txt("h"))},
		Rows:    [][]Cell{{blockCell()}},
	}).Doc()
	copy := Clone(orig)

	tbl := copy.Blocks[0].(Table)
	tbl.Caption[0] = Txt("MUTATED")
	cell := tbl.Rows[0][0]
	cell.Blocks[0] = Paragraph{Inlines: []Inline{Txt("MUTATED")}}
	cell.Inlines[0] = Txt("MUTATED")

	o := orig.Blocks[0].(Table)
	if got := o.Caption[0]; got != Txt("Caption") {
		t.Errorf("the original's caption changed to %#v", got)
	}
	oc := o.Rows[0][0]
	if got := oc.Blocks[0]; !reflect.DeepEqual(got, Paragraph{Inlines: []Inline{Txt("one")}}) {
		t.Errorf("the original cell's blocks changed to %#v", got)
	}
	if got := oc.Inlines[0]; got != Txt("one") {
		t.Errorf("the original cell's inlines changed to %#v", got)
	}
}

// countingVisitor records every node it is handed.
type countingVisitor struct{ seen []string }

func (v *countingVisitor) Enter(n any) bool {
	if t, ok := n.(Text); ok {
		v.seen = append(v.seen, t.Value)
	}
	return true
}
func (v *countingVisitor) Leave(any) {}

// TestWalkVisitsACellOnce is the one that keeps the duplication honest. A cell
// that carries its content in Blocks AND in Inlines would be visited twice by a
// walker that descends into both, so anything counting words -- or collecting
// links, or rewriting text -- would see every such cell's content doubled.
func TestWalkVisitsACellOnce(t *testing.T) {
	v := &countingVisitor{}
	Walk(New().Add(Table{Rows: [][]Cell{{blockCell()}}}).Doc(), v)
	want := []string{"one"} // from Blocks: the Paragraph's text. CodeBlock is not a Text.
	if !reflect.DeepEqual(v.seen, want) {
		t.Errorf("Walk saw %#v, want %#v", v.seen, want)
	}
}

// TestPlainTextReadsTheCaptionAndCountsACellOnce checks the same thing through the
// text extractor, which is what search and previews use.
func TestPlainTextReadsTheCaptionAndCountsACellOnce(t *testing.T) {
	got := PlainText(New().Add(Table{
		Caption: []Inline{Txt("Cap")},
		Rows:    [][]Cell{{blockCell()}},
	}).Doc())
	const want = "Cap one\ntwo"
	if got != want {
		t.Errorf("PlainText = %q, want %q", got, want)
	}
}
