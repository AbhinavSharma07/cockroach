// Copyright 2018 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cat

import (
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

// ColumnKind differentiates between different kinds of table columns.
type ColumnKind uint8

const (
	// Ordinary columns are "regular" table columns (including hidden columns
	// like `rowid`).
	Ordinary ColumnKind = iota
	// WriteOnly columns are mutation columns that have to be updated on writes
	// (inserts, updates, deletes) and cannot be otherwise accessed.
	WriteOnly
	// DeleteOnly columns are mutation columns that have to be updated only on
	// deletes and cannot be otherwise accessed.
	DeleteOnly
	// System columns are implicit columns that every physical table
	// contains. These columns can only be read from and must not be included
	// as part of mutations. They also cannot be part of the lax or key columns
	// for indexes. System columns are not members of any column family.
	System
	// Virtual columns are implicit columns that are used by inverted indexes (and
	// later, expression-based indexes).
	Virtual
)

// IsMutation is a convenience method that returns true if the column kind is
// a mutation column.
func (kind ColumnKind) IsMutation() bool {
	return kind == WriteOnly || kind == DeleteOnly
}

// IsSelectable is a convenience method that returns true if the column
// kind is a selectable column.
func (kind ColumnKind) IsSelectable() bool {
	return kind == Ordinary || kind == System
}

// Column is an interface to a table column, exposing only the information
// needed by the query optimizer.
type Column interface {
	// ColID is the unique, stable identifier for this column within its table.
	// Each new column in the table will be assigned a new ID that is different
	// than every column allocated before or after. This is true even if a column
	// is dropped and then re-added with the same name; the new column will have
	// a different ID. See the comment for StableID for more detail.
	//
	// Virtual columns don't have stable IDs; for these columns ColID() must not
	// be called.
	ColID() StableID

	// ColName returns the name of the column.
	ColName() tree.Name

	// DatumType returns the data type of the column.
	DatumType() *types.T

	// IsNullable returns true if the column is nullable.
	IsNullable() bool

	// IsHidden returns true if the column is hidden (e.g., there is always a
	// hidden column called rowid if there is no primary key on the table).
	IsHidden() bool

	// HasDefault returns true if the column has a default value. DefaultExprStr
	// will be set to the SQL expression string in that case.
	HasDefault() bool

	// DefaultExprStr is set to the SQL expression string that describes the
	// column's default value. It is used when the user does not provide a value
	// for the column when inserting a row. Default values cannot depend on other
	// columns.
	DefaultExprStr() string

	// IsComputed returns true if the column is a computed value. ComputedExprStr
	// will be set to the SQL expression string in that case.
	IsComputed() bool

	// ComputedExprStr is set to the SQL expression string that describes the
	// column's computed value. It is always used to provide the column's value
	// when inserting or updating a row. Computed values cannot depend on other
	// computed columns, but they can depend on all other columns, including
	// columns with default values.
	ComputedExprStr() string

	// InvertedSourceColumnOrdinal is implemented by virtual columns that are part
	// of inverted indexes. It returns the ordinal of the table column from which
	// the inverted column is derived.
	//
	// For example, if we have an inverted index on a JSON column `j`, the index
	// is on a virtual `j_inverted` column and calling
	// InvertedSourceColumnOrdinal() on `j_inverted` returns the ordinal of the
	// `j` column.
	//
	// Must not be called if this is not a virtual column.
	InvertedSourceColumnOrdinal() int
}

// IsMutationColumn is a convenience function that returns true if the column at
// the given ordinal position is a mutation column.
func IsMutationColumn(table Table, ord int) bool {
	return table.ColumnKind(ord).IsMutation()
}

// IsSystemColumn is a convenience function that returns true if the column at
// the given ordinal position is a system column.
func IsSystemColumn(table Table, ord int) bool {
	return table.ColumnKind(ord) == System
}

// IsSelectableColumn is a convenience function that returns true if the column
// at the given ordinal position is a selectable column.
func IsSelectableColumn(table Table, ord int) bool {
	return table.ColumnKind(ord).IsSelectable()
}

// IsVirtualColumn is a convenience function that returns true if the column at
// the given ordinal position is a virtual column.
func IsVirtualColumn(table Table, ord int) bool {
	return table.ColumnKind(ord) == Virtual
}
