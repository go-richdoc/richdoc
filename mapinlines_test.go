// Copyright (c) the go-richdoc authors.
// SPDX-License-Identifier: BSD-3-Clause

package richdoc

import (
	"fmt"
	"strings"
	"testing"
)

// ⛔ THE HAZARD THIS FUNCTION EXISTS TO REMOVE: a container the rewrite forgets
// keeps the un-rewritten node, and nothing is missing from the output to say
// so. This document puts one image in EVERY container the model has, and the
// test fails if a single one comes back un-rewritten — which is the test the
// two converters that each wrote this by hand did not have.
func TestEveryContainerIsRewritten(t *testing.T) {
	img := func(name string) Inline { return Image{URL: name} }
	doc := &Document{
		Meta: map[string]string{"title": "kept"},
		Blocks: []Block{
			Paragraph{Inlines: []Inline{img("paragraph")}},
			Heading{Level: 1, Inlines: []Inline{img("heading")}},
			BlockQuote{Blocks: []Block{Paragraph{Inlines: []Inline{img("quote")}}}},
			List{Items: []ListItem{{Blocks: []Block{
				Paragraph{Inlines: []Inline{img("list item")}},
			}}}},
			Table{
				Caption: []Inline{img("caption")},
				Header:  []Cell{{Inlines: []Inline{img("header cell")}}},
				Rows: [][]Cell{{{
					Inlines: []Inline{img("cell inlines")},
					Blocks:  []Block{Paragraph{Inlines: []Inline{img("cell blocks")}}},
				}}},
			},
			Paragraph{Inlines: []Inline{
				Emph{Inlines: []Inline{img("emph")}},
				Strong{Inlines: []Inline{img("strong")}},
				Strikethrough{Inlines: []Inline{img("strike")}},
				Link{URL: "/x", Inlines: []Inline{img("link")}},
				Anchor{ID: "a", Inlines: []Inline{img("anchor")}},
				CrossRef{Target: "t", Inlines: []Inline{img("crossref")}},
				Footnote{Blocks: []Block{Paragraph{Inlines: []Inline{img("footnote")}}}},
			}},
			// Blocks with no inline content at all: they must come through
			// untouched rather than be dropped.
			CodeBlock{Text: "code"},
			ThematicBreak{},
			MathBlock{TeX: "x^2"},
			RawBlock{Format: "svg", Text: "<svg/>"},
		},
	}

	var seen []string
	out := MapDocumentInlines(doc, func(in Inline) Inline {
		if i, ok := in.(Image); ok {
			seen = append(seen, i.URL)
			i.URL = "rewritten:" + i.URL
			return i
		}
		return in
	})

	want := []string{
		"paragraph", "heading", "quote", "list item", "caption", "header cell",
		"cell inlines", "cell blocks", "emph", "strong", "strike", "link",
		"anchor", "crossref", "footnote",
	}
	if len(seen) != len(want) {
		t.Errorf("f saw %d images, want %d: %v", len(seen), len(want), seen)
	}
	for _, w := range want {
		if !contains(seen, w) {
			t.Errorf("f never saw the image in the %s: a container is missing from the rewrite", w)
		}
	}

	// And every one of them really is rewritten in the OUTPUT, which is the
	// half a visitor cannot do.
	var left []string
	Walk(out, collector(func(n any) {
		if i, ok := n.(Image); ok && !strings.HasPrefix(i.URL, "rewritten:") {
			left = append(left, i.URL)
		}
	}))
	if len(left) > 0 {
		t.Errorf("%d image(s) came back un-rewritten: %v", len(left), left)
	}

	// The blocks that hold no inlines are still there.
	var code, breaks, maths, raws int
	Walk(out, collector(func(n any) {
		switch n.(type) {
		case CodeBlock:
			code++
		case ThematicBreak:
			breaks++
		case MathBlock:
			maths++
		case RawBlock:
			raws++
		}
	}))
	if code != 1 || breaks != 1 || maths != 1 || raws != 1 {
		t.Errorf("code=%d breaks=%d maths=%d raw=%d, want one of each", code, breaks, maths, raws)
	}
	if out.Meta["title"] != "kept" {
		t.Errorf("Meta = %v, want it copied", out.Meta)
	}
}

// ⛔ The original must not change. A converter that rewrote in place would
// surprise a caller holding the document for something else, and Clone exists
// precisely because that surprise is expensive.
func TestTheInputIsNotTouched(t *testing.T) {
	doc := &Document{Blocks: []Block{
		Paragraph{Inlines: []Inline{Image{URL: "before"}, Emph{Inlines: []Inline{Image{URL: "nested"}}}}},
		Table{Rows: [][]Cell{{{Inlines: []Inline{Image{URL: "cell"}}}}}},
		List{Items: []ListItem{{Blocks: []Block{Paragraph{Inlines: []Inline{Image{URL: "item"}}}}}}},
	}}
	_ = MapInlines(doc.Blocks, func(Inline) Inline { return Text{Value: "replaced"} })

	var urls []string
	Walk(doc, collector(func(n any) {
		if i, ok := n.(Image); ok {
			urls = append(urls, i.URL)
		}
	}))
	want := []string{"before", "nested", "cell", "item"}
	if fmt.Sprint(urls) != fmt.Sprint(want) {
		t.Errorf("the input now reads %v, want %v", urls, want)
	}
}

// f sees a container with its children ALREADY rewritten, which is what lets
// it replace a link by its own contents without visiting them twice.
func TestAContainerIsSeenAfterItsChildren(t *testing.T) {
	doc := []Block{Paragraph{Inlines: []Inline{
		Link{URL: "/x", Inlines: []Inline{Text{Value: "one"}, Text{Value: "two"}}},
	}}}
	var order []string
	out := MapInlines(doc, func(in Inline) Inline {
		switch v := in.(type) {
		case Text:
			order = append(order, "text:"+v.Value)
			return Text{Value: strings.ToUpper(v.Value)}
		case Link:
			order = append(order, fmt.Sprintf("link:%d children", len(v.Inlines)))
			// Unwrap the link: its children, already rewritten, stand alone.
			return Emph{Inlines: v.Inlines}
		}
		return in
	})
	if fmt.Sprint(order) != fmt.Sprint([]string{"text:one", "text:two", "link:2 children"}) {
		t.Errorf("order = %v, want the children before the container", order)
	}
	em, ok := out[0].(Paragraph).Inlines[0].(Emph)
	if !ok {
		t.Fatalf("Inlines[0] = %T, want the Emph that replaced the Link", out[0].(Paragraph).Inlines[0])
	}
	if got := em.Inlines[0].(Text).Value; got != "ONE" {
		t.Errorf("the unwrapped child reads %q, want the rewritten %q", got, "ONE")
	}
}

func TestNilAndEmpty(t *testing.T) {
	if got := MapInlines(nil, func(in Inline) Inline { return in }); got != nil {
		t.Errorf("MapInlines(nil) = %v, want nil", got)
	}
	if got := MapDocumentInlines(nil, func(in Inline) Inline { return in }); got != nil {
		t.Errorf("MapDocumentInlines(nil) = %v, want nil", got)
	}
	// A document with no metadata keeps none: an empty map would make a caller
	// testing for a title look twice.
	out := MapDocumentInlines(&Document{Blocks: []Block{Paragraph{}}}, func(in Inline) Inline { return in })
	if out.Meta != nil {
		t.Errorf("Meta = %v, want nil", out.Meta)
	}
	// Nil slices inside the containers stay nil rather than becoming empty
	// ones, so a rewritten tree compares equal to a hand-built one.
	out2 := MapInlines([]Block{
		Paragraph{}, List{}, Table{}, BlockQuote{},
		Paragraph{Inlines: []Inline{Emph{}, Footnote{}}},
	}, func(in Inline) Inline { return in })
	if p := out2[0].(Paragraph); p.Inlines != nil {
		t.Errorf("Paragraph.Inlines = %v, want nil", p.Inlines)
	}
	if l := out2[1].(List); l.Items != nil {
		t.Errorf("List.Items = %v, want nil", l.Items)
	}
	if tb := out2[2].(Table); tb.Rows != nil || tb.Header != nil || tb.Caption != nil {
		t.Errorf("Table came back with empty slices where it had nil")
	}
	if q := out2[3].(BlockQuote); q.Blocks != nil {
		t.Errorf("BlockQuote.Blocks = %v, want nil", q.Blocks)
	}
}

// collector adapts a function to Visitor for the assertions above.
type collector func(any)

func (c collector) Enter(n any) bool { c(n); return true }
func (c collector) Leave(any)        {}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
