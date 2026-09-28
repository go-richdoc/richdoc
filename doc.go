// Copyright (c) the go-richdoc authors.
// SPDX-License-Identifier: BSD-3-Clause

// Package richdoc defines a neutral, format-agnostic and widget-agnostic
// model for rich text documents.
//
// The model is a typed tree, not a generic attribute bag. A [Document] is an
// ordered slice of [Block] nodes; blocks and inlines are closed interface
// sets (each concrete type carries an unexported marker method), so consumers
// such as converters and editor widgets can exhaustively type-switch over
// them.
//
// The package is deliberately small and orthogonal. On top of the model it
// provides four utilities:
//
//   - [Walk] with a [Visitor] performs a depth-first traversal.
//   - [Builder] (via [New]) offers fluent, ergonomic construction.
//   - [PlainText] extracts the textual content of a document.
//   - [Clone] returns a deep copy.
//
// The package has no dependencies beyond the standard library and is safe to
// build with CGO disabled.
package richdoc

// Document is a rich text document: an ordered sequence of top-level blocks
// together with format-agnostic metadata (title, author, and similar).
//
// Meta is an optional, unstructured string map; converters decide how to map
// its keys onto their target format. A nil Meta is valid and means "no
// metadata".
type Document struct {
	Blocks []Block
	Meta   map[string]string
}

// Classes carry the presentational names an author attached to a node — reST's
// ":class:" option and ".. class::" directive, sphinx's ".. rst-class::", an
// HTML class attribute, a LaTeX \DUrole.
//
// They are the one deliberately UNSEMANTIC field in this model, and they are
// here because dropping them loses the author's own words: measured over
// go-richdoc/rst's 1564-document corpus, 187 class attributes in 57 files had
// nowhere to go, on paragraphs, tables, images, block quotes, lists and code.
// Every format in reach has somewhere to put them, and a model without them
// forces each converter to discard what it was given.
//
// What they are NOT: a general attribute bag. Nothing in this package
// interprets a class, no behaviour depends on one, and a converter whose format
// has no equivalent ignores the field. A node's MEANING stays in its type — the
// reason this model is a typed tree and not a map — and a class is a hint for
// whoever renders it.
//
// The slice is nil when the author wrote none, which is the overwhelming common
// case, so existing literals and Builder calls keep working unchanged.
