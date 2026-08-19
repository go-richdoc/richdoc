// Copyright (c) the go-richdoc authors.
// SPDX-License-Identifier: BSD-3-Clause

package richdoc

import (
	"reflect"
	"testing"
)

// fullDoc builds a document exercising every block and inline node type, with
// non-nil nested slices throughout. It is the backbone of the walk, plain-text
// and clone tests.
func fullDoc() *Document {
	return New().
		Meta("title", "T").
		Meta("author", "A").
		H(1, Txt("Title")).
		P(Bold(Txt("bold")), Txt(" and "), Italic(Txt("italic")), Strike(Txt("gone")),
			Mono("x"), Href("http://e", "t", Txt("link")), Img("i.png", "alt", "t"),
			InlineMath("a^2"), Br(), RawI("html", "<br>")).
		CodeBlock("go", "package main").
		Quote(Paragraph{Inlines: []Inline{Txt("quoted")}}).
		UList(true, Item(Paragraph{Inlines: []Inline{Txt("u1")}})).
		OList(3, false, Item(Paragraph{Inlines: []Inline{Txt("o1")}})).
		Table(
			[]Alignment{AlignLeft, AlignCenter, AlignRight, AlignDefault},
			[]Cell{Td(Txt("h1"))},
			[][]Cell{{Td(Txt("r1"))}},
		).
		HR().
		MathBlock(`\int x`).
		RawBlock("latex", `\newpage`).
		Add(Paragraph{Inlines: []Inline{Txt("added")}}).
		Doc()
}

// recordVisitor records the sequence of Enter/Leave calls and can prune.
type recordVisitor struct {
	enter []string
	leave []string
	skip  func(any) bool
}

func kind(n any) string { return reflect.TypeOf(n).String() }

func (r *recordVisitor) Enter(n any) bool {
	r.enter = append(r.enter, kind(n))
	if r.skip != nil {
		return !r.skip(n)
	}
	return true
}

func (r *recordVisitor) Leave(n any) { r.leave = append(r.leave, kind(n)) }

func TestWalkVisitsEveryNodeType(t *testing.T) {
	d := fullDoc()
	v := &recordVisitor{}
	Walk(d, v)

	if len(v.enter) != len(v.leave) {
		t.Fatalf("enter/leave mismatch: %d vs %d", len(v.enter), len(v.leave))
	}
	want := []string{
		"*richdoc.Document",
		"richdoc.Heading", "richdoc.Text",
		"richdoc.Paragraph",
		"richdoc.Strong", "richdoc.Text",
		"richdoc.Text",
		"richdoc.Emph", "richdoc.Text",
		"richdoc.Strikethrough", "richdoc.Text",
		"richdoc.Code",
		"richdoc.Link", "richdoc.Text",
		"richdoc.Image",
		"richdoc.Math",
		"richdoc.LineBreak",
		"richdoc.RawInline",
		"richdoc.CodeBlock",
		"richdoc.BlockQuote", "richdoc.Paragraph", "richdoc.Text",
		"richdoc.List", "richdoc.ListItem", "richdoc.Paragraph", "richdoc.Text",
		"richdoc.List", "richdoc.ListItem", "richdoc.Paragraph", "richdoc.Text",
		"richdoc.Table",
		"richdoc.Cell", "richdoc.Text",
		"richdoc.Cell", "richdoc.Text",
		"richdoc.ThematicBreak",
		"richdoc.MathBlock",
		"richdoc.RawBlock",
		"richdoc.Paragraph", "richdoc.Text",
	}
	if !reflect.DeepEqual(v.enter, want) {
		t.Fatalf("enter order:\n got %v\nwant %v", v.enter, want)
	}
}

func TestWalkNilDocument(t *testing.T) {
	v := &recordVisitor{}
	Walk(nil, v)
	if len(v.enter) != 0 || len(v.leave) != 0 {
		t.Fatalf("nil doc should not visit anything")
	}
}

func TestWalkEmptyDocument(t *testing.T) {
	v := &recordVisitor{}
	Walk(&Document{}, v)
	if !reflect.DeepEqual(v.enter, []string{"*richdoc.Document"}) {
		t.Fatalf("empty doc entered %v", v.enter)
	}
}

func TestWalkSkipPrunesChildren(t *testing.T) {
	d := New().P(Bold(Txt("x"))).Doc()

	// Prune at the document root: no blocks visited, but Leave still fires.
	v := &recordVisitor{skip: func(n any) bool { _, ok := n.(*Document); return ok }}
	Walk(d, v)
	if !reflect.DeepEqual(v.enter, []string{"*richdoc.Document"}) {
		t.Fatalf("root prune entered %v", v.enter)
	}
	if !reflect.DeepEqual(v.leave, []string{"*richdoc.Document"}) {
		t.Fatalf("root prune left %v", v.leave)
	}

	// Prune inside a block and an inline container: children skipped.
	v = &recordVisitor{skip: func(n any) bool {
		switch n.(type) {
		case Paragraph, Strong:
			return true
		}
		return false
	}}
	Walk(d, v)
	if !reflect.DeepEqual(v.enter, []string{"*richdoc.Document", "richdoc.Paragraph"}) {
		t.Fatalf("block prune entered %v", v.enter)
	}
}

func TestWalkSkipContainersIndividually(t *testing.T) {
	// Prune the ListItem and Cell containers to cover their skip branches.
	d := New().
		UList(false, Item(Paragraph{Inlines: []Inline{Txt("x")}})).
		Table(nil, []Cell{Td(Txt("h"))}, nil).
		Doc()
	v := &recordVisitor{skip: func(n any) bool {
		switch n.(type) {
		case ListItem, Cell:
			return true
		}
		return false
	}}
	Walk(d, v)
	want := []string{
		"*richdoc.Document",
		"richdoc.List", "richdoc.ListItem",
		"richdoc.Table", "richdoc.Cell",
	}
	if !reflect.DeepEqual(v.enter, want) {
		t.Fatalf("container prune entered %v", v.enter)
	}
}

func TestPlainText(t *testing.T) {
	got := PlainText(fullDoc())
	want := "Title\n" +
		"bold and italicgonexlink\n" +
		"package main\n" +
		"quoted\n" +
		"u1\n" +
		"o1\n" +
		"h1 r1\n" +
		"added"
	if got != want {
		t.Fatalf("PlainText:\n got %q\nwant %q", got, want)
	}
}

func TestPlainTextNil(t *testing.T) {
	if PlainText(nil) != "" {
		t.Fatalf("nil PlainText should be empty")
	}
}

func TestPlainTextEmptyBlocksSkipped(t *testing.T) {
	// A document whose only content is text-less blocks yields "".
	d := New().HR().MathBlock("x").RawBlock("html", "<hr>").
		P(Img("i", "", ""), Br()).
		UList(false, Item(ThematicBreak{})).
		Table(nil, []Cell{Td()}, [][]Cell{{Td()}}).
		Doc()
	if got := PlainText(d); got != "" {
		t.Fatalf("expected empty plain text, got %q", got)
	}
}

func TestCloneRoundTrip(t *testing.T) {
	orig := fullDoc()
	cl := Clone(orig)
	if !reflect.DeepEqual(orig, cl) {
		t.Fatalf("clone not equal to original")
	}

	// Mutating the clone must not affect the original.
	cl.Meta["title"] = "changed"
	cl.Blocks[0] = Paragraph{Inlines: []Inline{Txt("mutated")}}
	if orig.Meta["title"] != "T" {
		t.Fatalf("meta shared after clone")
	}
	if _, ok := orig.Blocks[0].(Heading); !ok {
		t.Fatalf("blocks slice shared after clone")
	}

	// Deep mutation of a nested slice must not leak either.
	cl2 := Clone(orig)
	tbl := cl2.Blocks[6].(Table)
	tbl.Rows[0][0] = Td(Txt("X"))
	tbl.Align[0] = AlignRight
	origTbl := orig.Blocks[6].(Table)
	if got := inlinesText(origTbl.Rows[0][0].Inlines); got != "r1" {
		t.Fatalf("table rows shared, got %q", got)
	}
	if origTbl.Align[0] != AlignLeft {
		t.Fatalf("table align shared")
	}
}

func TestCloneNil(t *testing.T) {
	if Clone(nil) != nil {
		t.Fatalf("Clone(nil) should be nil")
	}
}

func TestCloneNilFields(t *testing.T) {
	// A document with nil Meta and nil nested slices exercises every
	// nil-guard in the clone helpers.
	d := &Document{Blocks: []Block{
		Heading{Level: 1}, // nil Inlines
		BlockQuote{},      // nil Blocks
		List{},            // nil Items
		Table{},           // nil Align/Header/Rows
		// Emph/Strong with nil Inlines exercise the inline clone path.
		Paragraph{Inlines: []Inline{Emph{}, Strong{}}},
	}}

	cl := Clone(d)
	if cl.Meta != nil {
		t.Fatalf("nil meta should stay nil")
	}
	if !reflect.DeepEqual(d, cl) {
		t.Fatalf("clone with nil fields not equal")
	}

	// Direct coverage of the standalone nil guards.
	if cloneBlocks(nil) != nil || cloneInlines(nil) != nil ||
		cloneItems(nil) != nil || cloneCells(nil) != nil {
		t.Fatalf("clone helpers must return nil for nil input")
	}
}

func TestBuilderMetaInitAndReuse(t *testing.T) {
	// No metadata => nil Meta.
	if New().P(Txt("x")).Doc().Meta != nil {
		t.Fatalf("expected nil meta")
	}
	// Two keys exercise both the init and the reuse branch of Meta.
	m := New().Meta("a", "1").Meta("b", "2").Doc().Meta
	if m["a"] != "1" || m["b"] != "2" {
		t.Fatalf("meta not set: %v", m)
	}
}

func TestBuilderOListClampsStart(t *testing.T) {
	l := New().OList(0, false, Item(Paragraph{})).Doc().Blocks[0].(List)
	if l.Start != 1 || !l.Ordered {
		t.Fatalf("OList start not clamped: %+v", l)
	}
}
