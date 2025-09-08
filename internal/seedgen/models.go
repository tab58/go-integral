package seedgen

type TableSchema struct {
	OriginalTableName   string
	OriginalColumnNames []string
	TableName           string
	TableColumns        []TableSchemaColumn
	InputColumns        []TableSchemaColumn
	DependencyTables    map[string][]string
	InputToOutputMap    map[string]OutputMapData
}

type TableSchemaColumn struct {
	Name   string
	GoType string
}

type OutputMapData struct {
	ObjectName string
	FieldName  string
}

type GolangFile struct {
	Filename string
	Contents string
}
