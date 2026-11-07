// Copyright (c) the go-richdoc authors.
// SPDX-License-Identifier: BSD-3-Clause

package bombtest

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// fmtSprintf and keepAlive are named here so the recorder below reads as the
// small thing it is, without fmt and runtime appearing in the middle of it.
func fmtSprintf(f string, a ...any) string { return fmt.Sprintf(f, a...) }
func keepAlive(v any)                      { runtime.KeepAlive(v) }

// recorder is a [TB] that remembers instead of failing, so the refusals in
// [Run] can be tested rather than described.
//
// ⛔ Fatalf has to STOP, like the real one: testing.T's Fatalf calls
// runtime.Goexit, and a recorder that merely noted the message would let Run
// carry on into the very work the refusal exists to prevent — which would make
// the test pass while proving the opposite.
type recorder struct {
	fatal    []string
	errs     []string
	logs     []string
	cleanups []func()
}

type stopped struct{}

func (r *recorder) Helper() {}
func (r *recorder) Logf(f string, a ...any) {
	r.logs = append(r.logs, sprintf(f, a...))
}
func (r *recorder) Errorf(f string, a ...any) {
	r.errs = append(r.errs, sprintf(f, a...))
}
func (r *recorder) Fatalf(f string, a ...any) {
	r.fatal = append(r.fatal, sprintf(f, a...))
	panic(stopped{})
}
func (r *recorder) Cleanup(fn func()) { r.cleanups = append(r.cleanups, fn) }

func sprintf(f string, a ...any) string { return fmtSprintf(f, a...) }

// run drives Run against a recorder, catching the stop that Fatalf raises, and
// runs whatever cleanups were registered so the process-wide memory limit is
// restored exactly as it would be under a real test.
func run(b Bound, f func() error) (*recorder, Result) {
	r := &recorder{}
	var res Result
	func() {
		defer func() {
			if v := recover(); v != nil {
				if _, ok := v.(stopped); !ok {
					panic(v)
				}
			}
		}()
		res = Run(r, b, f)
	}()
	for i := len(r.cleanups) - 1; i >= 0; i-- {
		r.cleanups[i]()
	}
	return r, res
}

var refused = errors.New("refused")

// TestItRefusesABoundThatWasNotDeclared, and refuses it BEFORE the bomb runs.
//
// ⛔ Each case asserts that f NEVER RAN. A refusal that arrives after the
// allocation it failed to declare is not a gate, it is a report — and that
// distinction is the whole reason this package exists rather than a third copy
// of a comment.
func TestItRefusesABoundThatWasNotDeclared(t *testing.T) {
	for _, c := range []struct {
		name  string
		bound Bound
		says  string
	}{
		{"no reason", Bound{HeldAtMost: 1 << 20, AsksAtMost: 1 << 10}, "Why is empty"},
		{"no allocation ceiling", Bound{HeldAtMost: 1 << 20, Why: "x"}, "AsksAtMost is 0"},
		{"no held ceiling", Bound{AsksAtMost: 1 << 10, Why: "x"}, "HeldAtMost is 0"},
		{"nothing at all", Bound{}, "Why is empty"},
	} {
		t.Run(c.name, func(t *testing.T) {
			ran := false
			r, _ := run(c.bound, func() error { ran = true; return refused })

			if ran {
				t.Error("the bomb ran: the refusal has to arrive before the expense")
			}
			if len(r.fatal) != 1 {
				t.Fatalf("fatal messages: %v", r.fatal)
			}
			if !strings.Contains(r.fatal[0], c.says) {
				t.Errorf("the refusal does not say which field: %q", r.fatal[0])
			}
		})
	}
}

// TestItDoesNotFireOnADeclaredBound is the witness.
//
// ⛔ A GATE'S TEST HAS TO SHOW IT STAYS SHUT AS WELL AS THAT IT OPENS, or it
// proves only that the gate refuses everything — which is a gate nobody can
// use and which would be removed within a day.
func TestItDoesNotFireOnADeclaredBound(t *testing.T) {
	ran := false
	r, res := run(Bound{HeldAtMost: 1 << 30, AsksAtMost: 1 << 30, Why: "an ordinary bomb"},
		func() error { ran = true; return refused })

	if !ran {
		t.Error("the bomb did not run")
	}
	if len(r.fatal) != 0 || len(r.errs) != 0 {
		t.Errorf("a fully declared bound was refused: fatal=%v errs=%v", r.fatal, r.errs)
	}
	if res.Err == nil {
		t.Error("Result.Err is nil")
	}
	if res.Held == 0 {
		t.Error("Result.Held is 0: the process holds something, so this is not a reading")
	}
}

// TestItFailsWhenTheInputWasAccepted: the assertion that cannot be switched off.
func TestItFailsWhenTheInputWasAccepted(t *testing.T) {
	r, _ := run(Bound{HeldAtMost: 1 << 30, AsksAtMost: 1 << 30, Why: "a bomb that is not one"},
		func() error { return nil })

	if len(r.fatal) != 1 || !strings.Contains(r.fatal[0], "ACCEPTED") {
		t.Fatalf("an accepted input did not fail the test: %v", r.fatal)
	}
	if !strings.Contains(r.fatal[0], "a bomb that is not one") {
		t.Errorf("the failure does not name what the bound protects: %q", r.fatal[0])
	}
}

// TestItNamesTheAllocationItDidNotDeclare.
//
// ⛔ The planted positive for the figure that NAMES a regression: a bomb that
// asks for a megabyte against a ceiling of one byte. Without this the
// AsksAtMost check is a claim that a comparison compiles.
func TestItNamesTheAllocationItDidNotDeclare(t *testing.T) {
	r, res := run(Bound{HeldAtMost: 1 << 30, AsksAtMost: 1, Why: "a bomb that allocates"},
		func() error {
			var kept [][]byte
			for range 16 {
				kept = append(kept, make([]byte, 1<<16))
			}
			keepAlive(kept)
			return refused
		})

	if len(r.errs) == 0 {
		t.Fatalf("asking for %d bytes against a ceiling of 1 was accepted", res.Asked)
	}
	if !strings.Contains(r.errs[0], "asked the allocator for") {
		t.Errorf("the failure does not name the figure: %q", r.errs[0])
	}
	if res.Asked < 1<<20 {
		t.Errorf("Result.Asked is %d for a megabyte of allocation: TotalAlloc's "+
			"difference is not being read", res.Asked)
	}
}

// TestItNamesTheMemoryItHeldPastTheCap is the planted positive for the held
// check, and it cost nothing to write: the process already holds megabytes, so
// a cap of one byte is past it before the bomb allocates anything.
//
// ⛔ IT EXISTS BECAUSE THE COVERAGE GATE ASKED FOR IT. Without it `Run` sat at
// 96.3% with exactly this branch unrun — which is the "a cap nobody has seen
// trip is not a cap" problem showing up as a number instead of as something
// somebody has to remember. That is the argument for the gate being code.
func TestItNamesTheMemoryItHeldPastTheCap(t *testing.T) {
	r, res := run(Bound{HeldAtMost: 1, AsksAtMost: 1 << 30, Why: "a cap of one byte"},
		func() error { return refused })

	if len(r.errs) == 0 {
		t.Fatalf("the process holds %d bytes and a cap of 1 was not reported", res.Held)
	}
	if !strings.Contains(r.errs[0], "the process holds") {
		t.Errorf("the failure does not name the figure: %q", r.errs[0])
	}
	if !strings.Contains(r.errs[0], "a cap of one byte") {
		t.Errorf("the failure does not name what the bound protects: %q", r.errs[0])
	}
}

// TestUnboundedStampsAndSilencesNothingElse.
//
// ⛔ A derogation that quiets the gate becomes the normal route. This one drops
// exactly one check, stamps every run it is used on, and leaves the refusal
// assertion and the allocation ceiling in force — asserted here, because
// "it only drops one thing" is the kind of claim that stops being true.
func TestUnboundedStampsAndSilencesNothingElse(t *testing.T) {
	// It is accepted without HeldAtMost...
	r, _ := run(Bound{AsksAtMost: 1 << 30, Why: "one enormous request", Unbounded: true},
		func() error { return refused })
	if len(r.fatal) != 0 {
		t.Fatalf("Unbounded with no HeldAtMost was refused: %v", r.fatal)
	}

	// ...and it stamps.
	stamped := false
	for _, l := range r.logs {
		if strings.Contains(l, "⛔ UNBOUNDED") && strings.Contains(l, "one enormous request") {
			stamped = true
		}
	}
	if !stamped {
		t.Errorf("Unbounded did not stamp the output: %v", r.logs)
	}

	// The refusal assertion still applies.
	r2, _ := run(Bound{AsksAtMost: 1 << 30, Why: "x", Unbounded: true},
		func() error { return nil })
	if len(r2.fatal) == 0 {
		t.Error("Unbounded silenced the refusal assertion")
	}

	// And so does the allocation ceiling.
	r3, _ := run(Bound{AsksAtMost: 1, Why: "x", Unbounded: true},
		func() error {
			keepAlive(make([]byte, 1<<20))
			return refused
		})
	if len(r3.errs) == 0 {
		t.Error("Unbounded silenced the allocation ceiling")
	}
}

// TestTheStringSaysBothFigures, since a test that logs it is logging the whole
// measurement.
func TestTheStringSaysBothFigures(t *testing.T) {
	s := Result{Held: 7, Asked: 11, Err: refused}.String()
	for _, want := range []string{"asked for 11", "holds 7", "refused"} {
		if !strings.Contains(s, want) {
			t.Errorf("%q is missing %q", s, want)
		}
	}
}
