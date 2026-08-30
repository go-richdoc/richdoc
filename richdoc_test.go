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

// TestCloneCellSpanFields guards against cloneCells rebuilding each Cell
// from just its Inlines field — an easy mistake once Cell gains fields
// beyond Inlines, and one this test would have caught: a first version of
// ColSpan/RowSpan support did exactly that, silently dropping both on
// every Clone.
func TestCloneCellSpanFields(t *testing.T) {
	orig := New().Table(nil, nil, [][]Cell{{{Inlines: []Inline{Txt("x")}, ColSpan: 2, RowSpan: 3}}}).Doc()
	cl := Clone(orig)
	cell := cl.Blocks[0].(Table).Rows[0][0]
	if cell.ColSpan != 2 || cell.RowSpan != 3 {
		t.Fatalf("clone dropped span fields: got ColSpan=%d RowSpan=%d, want 2 and 3", cell.ColSpan, cell.RowSpan)
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

// refDoc builds a document exercising the v0.2.0 additions (Footnote, Anchor,
// CrossRef and Heading.ID) with non-nil nested slices throughout.
func refDoc() *Document {
	return New().
		Add(Heading{Level: 2, ID: "sec", Inlines: []Inline{Txt("Sec")}}).
		P(
			Txt("body"),
			Note(Paragraph{Inlines: []Inline{Txt("note")}}),
			Mark("lbl", Txt("here")),
			Ref("sec", Txt("2")),
			Cite("knuth", Txt("K")),
		).
		Doc()
}

func TestNewNodeHelpersMatchLiterals(t *testing.T) {
	// The helpers must build exactly the same values as the struct literals.
	if got, want := Note(Paragraph{}), (Footnote{Blocks: []Block{Paragraph{}}}); !reflect.DeepEqual(got, want) {
		t.Fatalf("Note: got %+v want %+v", got, want)
	}
	if got, want := Mark("id", Txt("x")), (Anchor{ID: "id", Inlines: []Inline{Txt("x")}}); !reflect.DeepEqual(got, want) {
		t.Fatalf("Mark: got %+v want %+v", got, want)
	}
	// A point Anchor may carry no inlines.
	if got := Mark("id"); got.ID != "id" || got.Inlines != nil {
		t.Fatalf("point Mark: got %+v", got)
	}
	if got, want := Ref("t", Txt("x")), (CrossRef{Target: "t", Kind: RefLabel, Inlines: []Inline{Txt("x")}}); !reflect.DeepEqual(got, want) {
		t.Fatalf("Ref: got %+v want %+v", got, want)
	}
	if got, want := Cite("t", Txt("x")), (CrossRef{Target: "t", Kind: RefCite, Inlines: []Inline{Txt("x")}}); !reflect.DeepEqual(got, want) {
		t.Fatalf("Cite: got %+v want %+v", got, want)
	}
	if RefLabel != 0 || RefCite != 1 {
		t.Fatalf("RefKind values changed: RefLabel=%d RefCite=%d", RefLabel, RefCite)
	}
}

func TestWalkVisitsNewNodes(t *testing.T) {
	v := &recordVisitor{}
	Walk(refDoc(), v)

	if len(v.enter) != len(v.leave) {
		t.Fatalf("enter/leave mismatch: %d vs %d", len(v.enter), len(v.leave))
	}
	want := []string{
		"*richdoc.Document",
		"richdoc.Heading", "richdoc.Text",
		"richdoc.Paragraph",
		"richdoc.Text",
		// Footnote descends into its block-level body.
		"richdoc.Footnote", "richdoc.Paragraph", "richdoc.Text",
		// Anchor descends into its inline content.
		"richdoc.Anchor", "richdoc.Text",
		// CrossRef (RefLabel then RefCite) descends into its visible inlines.
		"richdoc.CrossRef", "richdoc.Text",
		"richdoc.CrossRef", "richdoc.Text",
	}
	if !reflect.DeepEqual(v.enter, want) {
		t.Fatalf("enter order:\n got %v\nwant %v", v.enter, want)
	}
}

func TestWalkSkipNewNodesPrunesChildren(t *testing.T) {
	// Pruning each new node skips its children but still fires Leave.
	v := &recordVisitor{skip: func(n any) bool {
		switch n.(type) {
		case Footnote, Anchor, CrossRef:
			return true
		}
		return false
	}}
	Walk(refDoc(), v)
	want := []string{
		"*richdoc.Document",
		"richdoc.Heading", "richdoc.Text",
		"richdoc.Paragraph",
		"richdoc.Text",
		"richdoc.Footnote",
		"richdoc.Anchor",
		"richdoc.CrossRef",
		"richdoc.CrossRef",
	}
	if !reflect.DeepEqual(v.enter, want) {
		t.Fatalf("pruned enter order:\n got %v\nwant %v", v.enter, want)
	}
}

func TestPlainTextNewNodes(t *testing.T) {
	// Footnote body text is included inline where the note occurs; Anchor and
	// CrossRef emit their visible inlines but never their identifiers.
	got := PlainText(refDoc())
	want := "Sec\n" + "bodynotehere2K"
	if got != want {
		t.Fatalf("PlainText:\n got %q\nwant %q", got, want)
	}

	// A point Anchor and empty-text CrossRefs contribute nothing, and their
	// ids/targets never leak into the text.
	d := New().P(Mark("lbl"), Ref("sec"), Cite("knuth")).Doc()
	if got := PlainText(d); got != "" {
		t.Fatalf("expected empty plain text, got %q", got)
	}
}

func TestCloneNewNodes(t *testing.T) {
	orig := refDoc()
	cl := Clone(orig)
	if !reflect.DeepEqual(orig, cl) {
		t.Fatalf("clone not equal to original")
	}

	// Heading.ID round-trips through Clone.
	if cl.Blocks[0].(Heading).ID != "sec" {
		t.Fatalf("Heading.ID not cloned: %+v", cl.Blocks[0])
	}

	// Deep-mutating the clone's new nodes must not touch the original.
	p := cl.Blocks[1].(Paragraph)
	p.Inlines[1].(Footnote).Blocks[0] = Paragraph{Inlines: []Inline{Txt("MUT")}}
	p.Inlines[2].(Anchor).Inlines[0] = Txt("MUT")
	p.Inlines[3].(CrossRef).Inlines[0] = Txt("MUT")

	op := orig.Blocks[1].(Paragraph)
	if got := blocksText(op.Inlines[1].(Footnote).Blocks); got != "note" {
		t.Fatalf("footnote blocks shared, got %q", got)
	}
	if got := inlinesText(op.Inlines[2].(Anchor).Inlines); got != "here" {
		t.Fatalf("anchor inlines shared, got %q", got)
	}
	if got := inlinesText(op.Inlines[3].(CrossRef).Inlines); got != "2" {
		t.Fatalf("crossref inlines shared, got %q", got)
	}

	// The nil-slice paths of the new nodes clone cleanly too.
	nd := &Document{Blocks: []Block{Paragraph{Inlines: []Inline{
		Footnote{}, Anchor{ID: "a"}, CrossRef{Target: "t", Kind: RefCite},
	}}}}
	if !reflect.DeepEqual(nd, Clone(nd)) {
		t.Fatalf("clone with nil new-node slices not equal")
	}
}

// TestBackwardCompatUnchanged proves the v0.2.0 additions are non-breaking: a
// document built entirely from v0.1.0 nodes walks, plain-texts and clones
// byte-for-byte identically to before the additions.
func TestBackwardCompatUnchanged(t *testing.T) {
	// Walk order for the legacy fullDoc is unchanged.
	v := &recordVisitor{}
	Walk(fullDoc(), v)
	wantWalk := []string{
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
	if !reflect.DeepEqual(v.enter, wantWalk) {
		t.Fatalf("legacy walk order changed:\n got %v\nwant %v", v.enter, wantWalk)
	}

	// PlainText for the legacy document is unchanged.
	wantText := "Title\n" +
		"bold and italicgonexlink\n" +
		"package main\n" +
		"quoted\n" +
		"u1\n" +
		"o1\n" +
		"h1 r1\n" +
		"added"
	if got := PlainText(fullDoc()); got != wantText {
		t.Fatalf("legacy PlainText changed:\n got %q\nwant %q", got, wantText)
	}

	// Clone of the legacy document is unchanged and independent.
	orig := fullDoc()
	if !reflect.DeepEqual(orig, Clone(orig)) {
		t.Fatalf("legacy clone not equal to original")
	}

	// A legacy Heading carries an empty ID, i.e. the field defaults are inert.
	if fullDoc().Blocks[0].(Heading).ID != "" {
		t.Fatalf("legacy Heading gained a non-empty ID")
	}
}
