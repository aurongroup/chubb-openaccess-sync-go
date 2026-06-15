package model

type RowObject interface {
	ToRow() []string
}

type RowObjectCache interface {
	RowHeader() []string
	GetRowItems() []RowObject
}
