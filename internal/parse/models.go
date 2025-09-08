package parse

type Table struct {
	Name        string            `json:"name"`
	Columns     []Column          `json:"columns"`
	Constraints []TableConstraint `json:"table_constraints"`
}

type Column struct {
	Name        string             `json:"name"`
	DataType    string             `json:"data_type"`
	Constraints []ColumnConstraint `json:"constraints"`
}

type ColumnConstraint struct {
	Type            ConstraintInfoType `json:"type"`
	ExpressionValue string             `json:"expression_value"`
}

type TableConstraint struct {
	Type       ConstraintInfoType `json:"type"`
	Constraint ConstraintInfo     `json:"constraint"`
}

type ConstraintInfoType string

const (
	ConstraintInfoTypeUnknown    ConstraintInfoType = "unknown"
	ConstraintInfoTypeForeignKey ConstraintInfoType = "foreign_key"
	ConstraintInfoTypePrimaryKey ConstraintInfoType = "primary_key"
	ConstraintInfoTypeUnique     ConstraintInfoType = "unique"
	ConstraintInfoTypeDefault    ConstraintInfoType = "default"
	ConstraintInfoTypeNotNull    ConstraintInfoType = "not_null"
)

type ConstraintInfo interface {
	ConstraintType() ConstraintInfoType
}

type PrimaryKeyConstraintInfo struct {
	ColumnNames []string `json:"column_names"`
}

func (c *PrimaryKeyConstraintInfo) ConstraintType() ConstraintInfoType {
	return ConstraintInfoTypePrimaryKey
}

type ForeignKeyConstraintInfo struct {
	TableColumnName      string `json:"table_column_name"`
	ForeignKeyTableName  string `json:"foreign_key_table_name"`
	ForeignKeyColumnName string `json:"foreign_key_column_name"`
}

func (c *ForeignKeyConstraintInfo) ConstraintType() ConstraintInfoType {
	return ConstraintInfoTypeForeignKey
}
