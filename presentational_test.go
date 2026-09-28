package richdoc_test

import (
	"reflect"
	"testing"

	"github.com/go-richdoc/richdoc"
)

// TestCloneDoesNotShareClasses is the only thing about these new fields that can
// go silently wrong. Clone promises the copy "shares no mutable state (slices or
// maps) with the original", and a []string is mutable state: without a copy,
// appending to one document's classes appends to the other's, and nothing reports
// it until somebody mutates a copy.
func TestCloneDoesNotShareClasses(t *testing.T) {
	orig := &richdoc.Document{Blocks: []richdoc.Block{
		richdoc.Paragraph{Inlines: []richdoc.Inline{richdoc.Text{Value: "x"}}, Classes: []string{"a"}},
		richdoc.List{Items: []richdoc.ListItem{{}}, Classes: []string{"b"}},
		richdoc.BlockQuote{Classes: []string{"c"}},
		richdoc.Table{Classes: []string{"d"}},
		richdoc.CodeBlock{Text: "t", Classes: []string{"e"}},
		richdoc.Paragraph{Inlines: []richdoc.Inline{
			richdoc.Code{Value: "v", Classes: []string{"f"}},
			richdoc.Image{URL: "u", Classes: []string{"g"}},
		}},
	}}
	copyOf := richdoc.Clone(orig)
	if !reflect.DeepEqual(orig, copyOf) {
		t.Fatalf("Clone is not equal to its original:\n%#v\n%#v", orig, copyOf)
	}
	// Mutate every cloned Classes slice in place. Sharing shows up as the
	// ORIGINAL changing.
	before := richdoc.Clone(orig)
	mutate(copyOf)
	if !reflect.DeepEqual(orig, before) {
		t.Errorf("mutating the clone changed the original:\n got %#v\nwant %#v", orig, before)
	}
}

// mutate overwrites the first element of every Classes slice it can reach.
func mutate(d *richdoc.Document) {
	for i, b := range d.Blocks {
		switch n := b.(type) {
		case richdoc.Paragraph:
			poke(n.Classes)
			for _, in := range n.Inlines {
				switch v := in.(type) {
				case richdoc.Code:
					poke(v.Classes)
				case richdoc.Image:
					poke(v.Classes)
				}
			}
		case richdoc.List:
			poke(n.Classes)
		case richdoc.BlockQuote:
			poke(n.Classes)
		case richdoc.Table:
			poke(n.Classes)
		case richdoc.CodeBlock:
			poke(n.Classes)
		}
		d.Blocks[i] = b
	}
}

func poke(classes []string) {
	if len(classes) > 0 {
		classes[0] = "MUTATED"
	}
}

// TestImageSizeFieldsAreZeroByDefault pins the compatibility promise: an existing
// literal or Builder call that names none of the new fields keeps working and
// says "not given" for every one of them. Scale uses 0 for that, a zero-percent
// image being meaningless, so no pointer is needed.
func TestImageSizeFieldsAreZeroByDefault(t *testing.T) {
	img := richdoc.Image{URL: "u", Alt: "a"}
	if img.Width != "" || img.Height != "" || img.Scale != 0 || img.Align != richdoc.AlignDefault || img.Classes != nil {
		t.Errorf("a plain Image should carry no size or class: %#v", img)
	}
}

// TestClassesSurviveWalk guards the other utility that walks every node: Walk
// must not lose or reorder what it visits now that nodes carry one more field.
// A field is not a child, so nothing here should change -- which is exactly what
// makes this worth checking once rather than assuming.
func TestClassesSurviveWalk(t *testing.T) {
	d := &richdoc.Document{Blocks: []richdoc.Block{
		richdoc.Paragraph{Inlines: []richdoc.Inline{richdoc.Text{Value: "x"}}, Classes: []string{"a"}},
	}}
	seen := 0
	v := &classVisitor{t: t, seen: &seen}
	richdoc.Walk(d, v)
	if seen != 1 {
		t.Errorf("visited the paragraph %d times, want 1", seen)
	}
}

// classVisitor counts the paragraphs Walk hands over and checks each still
// carries its classes.
type classVisitor struct {
	t    *testing.T
	seen *int
}

func (v *classVisitor) Enter(n any) bool {
	if p, ok := n.(richdoc.Paragraph); ok {
		*v.seen++
		if !reflect.DeepEqual(p.Classes, []string{"a"}) {
			v.t.Errorf("Walk handed over Classes = %#v", p.Classes)
		}
	}
	return true
}

func (v *classVisitor) Leave(any) {}
