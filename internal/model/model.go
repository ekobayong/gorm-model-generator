package model

// Table represents a database table
type Table struct {
	Name        string
	GoName      string // PascalCase name for Go struct
	Columns     []Column
	ForeignKeys []ForeignKey
	Relations   []Relation
}

// Column represents a database column
type Column struct {
	Name         string
	GoName       string // PascalCase name for Go field
	DataType     string
	IsNullable   bool
	IsPrimaryKey bool
	GoType       string
	GormTag      string
	JsonTag      string
	Comment      string
}

// ForeignKey represents a foreign key constraint
type ForeignKey struct {
	ConstraintName string
	ColumnName     string
	RefTableName   string
	RefColumnName  string
	OnDelete       string
	OnUpdate       string
}

// Relation represents a Go struct relationship (HasMany, BelongsTo)
type Relation struct {
	Name       string // Field name in struct e.g. "Users"
	Type       string // HasMany, BelongsTo, HasOne, ManyToMany
	OtherType  string // Type of the other struct e.g. "User"
	ForeignKey string // Field name of FK e.g. "CompanyID"
	References string // Field name of Reference e.g. "ID"
}
