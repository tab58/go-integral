package seedgen

type RawTableSchema struct {
	TableName        string
	TablePrimaryKey  []string
	TableColumns     []RawTableSchemaColumn
	InputColumns     []RawTableSchemaColumn
	DependencyTables map[string][]string
	InputToOutputMap map[string]OutputMapData
}
type RawTableSchemaColumn struct {
	Name   string
	GoType string
}

type TableSchema struct {
	TableName          SQLGolangStringValue
	SQLTablePrimaryKey []SQLGolangStringValue
	TableColumns       []TableSchemaColumn
	SQLColumnNames     []SQLGolangStringValue
	RecordInputColumns []TableSchemaColumn
	DependencyTables   []DependencyTable
	InputToOutputMap   map[string]OutputMapData
}

type TableSchemaColumn struct {
	Name   SQLGolangStringValue
	GoType string
}

type SQLGolangStringValue struct {
	SQL    string
	Golang string
}

type DependencyTable struct {
	GolangTableName string
	InputRecordName string
	ColumnNames     []string
}

type OutputMapData struct {
	ObjectName string
	FieldName  string
}

type GolangFile struct {
	Filename string
	Contents string
}
