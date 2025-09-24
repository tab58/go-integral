package nodes

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

// ----------

type ForeignKeyConstraintInfo struct {
	TableColumnName      string `json:"table_column_name"`
	ForeignKeyTableName  string `json:"foreign_key_table_name"`
	ForeignKeyColumnName string `json:"foreign_key_column_name"`
}

func (c *ForeignKeyConstraintInfo) ConstraintType() ConstraintInfoType {
	return ConstraintInfoTypeForeignKey
}

// ----------

type PrimaryKeyConstraintInfo struct {
	ColumnNames []string `json:"column_names"`
}

func (c *PrimaryKeyConstraintInfo) ConstraintType() ConstraintInfoType {
	return ConstraintInfoTypePrimaryKey
}

// ----------
