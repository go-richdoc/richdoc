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
