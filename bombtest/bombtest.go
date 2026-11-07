// Copyright (c) the go-richdoc authors.
// SPDX-License-Identifier: BSD-3-Clause

// Package bombtest runs a bomb test under a declared memory bound and refuses
// to run one that has not declared it.
//
// # Why this is a refusal and not a comment
//
// ⛔ A BOMB TEST IS A TEST THAT A HOSTILE INPUT IS REFUSED, and it is written by
// feeding the reader something enormous. Three mistakes are easy to make in
// that shape, and every one of them has been made in this family — twice each,
// with the lesson written down in between. A note and an optional helper fail
// the same way: both assume somebody remembers to consult them. So the
// questions live here as arguments that have NO DEFAULT, and the refusal
// arrives before the bomb runs rather than after it has allocated.
//
// The three:
//
//  1. A bomb test that forgets to assert the REFUSAL passes just as happily on a
//     reader that accepted the document and built a gigabyte of nodes — which is
//     the failure it exists to catch. [Run] asserts it, always, and there is no
//     way to ask it not to.
//
//  2. A test that bounds nothing can take the machine with it. A sibling
//     project's test asked for about 40 GB and OOM-killed a CI runner, while
//     darwin overcommitted lazily and hid it on the machine it was written on.
//     [Bound.HeldAtMost] is required.
//
//  3. A memory figure measured as a DELTA of [runtime.MemStats.Sys] is
//     order-dependent, because Sys never shrinks: the same bomb measured
//     94 736 400 bytes run first and 40 386 560 run second, and a self-check
//     for exactly this read 164 292 904 − 164 292 904 = 0 and passed alone
//     while failing last. This package does not expose a delta of Sys at all.
//     [Result.Held] is absolute, and [Result.Asked] is the delta of
//     TotalAlloc, which is cumulative and therefore sound.
//
// # Why two bounds and not one
//
// ⛔ A HELD-ONLY BOUND PROTECTS BY AVOIDANCE AND HIDES WHAT IT AVOIDS.
// GOMEMLIMIT — [debug.SetMemoryLimit], which [Run] sets — is a GC TARGET, not
// an allocation refusal. A reader whose ceiling regressed does not reserve more
// memory under it: it makes the collector work harder, so the failure arrives
// as a test that takes minutes, and a slow test reads as a slow machine rather
// than as a broken bound.
//
// So [Bound.AsksAtMost] is the one that NAMES a regression. Bytes handed out
// cannot be collected away, so it fires at once and says both numbers.
// Measured in go-richdoc/html: refusing one 3.2 MB document asks for
// 96 063 696 bytes through a ceiling checked while building and 7 096 through
// one checked over the source, so a tight AsksAtMost on the second names a
// four-order regression the moment it appears.
//
// ⛔ And the hard version of HeldAtMost — `ulimit -v` — is a NO-OP ON DARWIN
// (RLIMIT_AS is ignored there), so a shell guard protects CI and nothing on the
// machine where a bomb gets written, which is backwards. Hence the Go-side
// limit, and hence [Result.Held] being checked rather than trusted.
package bombtest

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// TB is the part of [testing.TB] this package uses. *testing.T satisfies it,
// and a test of this package can satisfy it too — which is how the refusals
// below are themselves tested.
type TB interface {
	Helper()
	Logf(format string, args ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	Cleanup(func())
}

// A Bound is what a bomb test has to declare before it is allowed to run.
//
// ⛔ THERE ARE NO DEFAULTS. A default here would be a silent assumption about a
// question the caller is the only one who can answer: how much this particular
// refusal legitimately costs. A zero field is refused by [Run] before the bomb
// runs, which is the whole point — the refusal has to arrive before the
// expense, not after 159 MB has been allocated.
type Bound struct {
	// HeldAtMost is how much the process may hold from the operating system
	// while the bomb runs, in bytes. Required.
	HeldAtMost uint64

	// AsksAtMost is how much the call may ask the allocator for, in bytes.
	// Required — and this is the figure that names a regression.
	AsksAtMost uint64

	// Why names, in one line, what this bound protects. Required, because a
	// failure that says only "past the ceiling" leaves the next reader to
	// rediscover which property broke.
	Why string

	// Unbounded drops the HeldAtMost check for a bomb that genuinely cannot be
	// run inside a cap — one enormous allocation, which a soft limit cannot
	// hold anyway and only RLIMIT_AS could.
	//
	// ⛔ IT STAMPS THE OUTPUT AND SILENCES NOTHING ELSE. The refusal assertion
	// and AsksAtMost still apply, and every run logs ⛔ UNBOUNDED with Why, so
	// a derogation cannot become the quiet normal route.
	Unbounded bool
}

// A Result is what the bomb cost.
//
// ⛔ Held is ABSOLUTE and Asked is a DELTA, and that asymmetry is the point:
// Sys never shrinks, so only its absolute value means anything, while
// TotalAlloc is cumulative, so only its difference does. Exposing the wrong one
// of each is the mistake this package exists to make unavailable.
type Result struct {
	// Held is runtime.MemStats.Sys after the bomb: what the process holds from
	// the operating system, which is what an OOM killer counts.
	Held uint64

	// Asked is how many bytes the bomb handed out: the difference of
	// TotalAlloc, which no amount of collection hides.
	Asked uint64

	// Err is what the bomb returned. Never nil: [Run] fails the test otherwise.
	Err error
}

// Run runs f under b and asserts that f refused, that it asked for no more than
// b.AsksAtMost, and — unless b.Unbounded — that the process held no more than
// b.HeldAtMost.
//
// It refuses, before running f at all, a Bound that has not been filled in.
func Run(t TB, b Bound, f func() error) Result {
	t.Helper()

	// ⛔ BEFORE THE EXPENSE. Everything below this block costs whatever the
	// bomb costs; everything in it costs nothing, so an incomplete Bound fails
	// in microseconds rather than after the allocation it failed to declare.
	if b.Why == "" {
		t.Fatalf("bombtest: Bound.Why is empty: name what this bound protects, " +
			"or a failure here will say only that a number was too big")
	}
	if b.AsksAtMost == 0 {
		t.Fatalf("bombtest: Bound.AsksAtMost is 0 for %q: it is required and has no "+
			"default, because how much this refusal legitimately asks for is the one "+
			"thing only its author knows — and it is the figure that names a regression",
			b.Why)
	}
	if b.HeldAtMost == 0 && !b.Unbounded {
		t.Fatalf("bombtest: Bound.HeldAtMost is 0 for %q: it is required and has no "+
			"default. Set it, or set Unbounded with a reason if this bomb cannot run "+
			"inside a cap", b.Why)
	}

	if !b.Unbounded {
		// Process-wide, so it is restored even if f panics. No test that sets
		// it may run in parallel with another.
		previous := debug.SetMemoryLimit(int64(b.HeldAtMost))
		t.Cleanup(func() { debug.SetMemoryLimit(previous) })
	}

	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	err := f()

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	r := Result{Held: after.Sys, Asked: after.TotalAlloc - before.TotalAlloc, Err: err}

	if b.Unbounded {
		t.Logf("⛔ UNBOUNDED (%s): the held-memory check is off; asked for %d bytes",
			b.Why, r.Asked)
	}

	// ⛔ A BOMB TEST ASSERTS A REFUSAL, and this is not optional. Without it the
	// test passes on a reader that accepted the document, which is the defect.
	if err == nil {
		t.Fatalf("bombtest: %s: the input was ACCEPTED. A bomb test that passes on an "+
			"accepted document measures nothing", b.Why)
	}

	if r.Asked > b.AsksAtMost {
		t.Errorf("bombtest: %s: refusing asked the allocator for %d bytes, past the "+
			"%d declared: the bound is no longer refusing before the work",
			b.Why, r.Asked, b.AsksAtMost)
	}
	if !b.Unbounded && r.Held > b.HeldAtMost {
		t.Errorf("bombtest: %s: the process holds %d bytes, past the %d declared: the "+
			"soft limit did not hold, so this figure was taken under memory pressure",
			b.Why, r.Held, b.HeldAtMost)
	}
	return r
}

// String is the one-line report of a bomb's cost, for a test that wants to log
// what it measured next to what it declared.
func (r Result) String() string {
	return fmt.Sprintf("refused (%v): asked for %d bytes, process holds %d",
		r.Err, r.Asked, r.Held)
}
