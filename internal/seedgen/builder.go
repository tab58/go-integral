package seedgen

import (
	"errors"
	"go-integral/internal/parse"
	"go-integral/internal/utils"
	"slices"
	"strings"

	"github.com/iancoleman/strcase"
	pg_query "github.com/pganalyze/pg_query_go/v6"
)

type Builder struct {
	sortedTables []parse.Table
}

func NewFromSQLSchema(sqlSchema string) (*Builder, error) {
	pgQuery, err := pg_query.Parse(sqlSchema)
	if err != nil {
		return nil, errors.New("failed to parse PostgreSQL schema source: " + err.Error())
	}

	mgr := parse.NewEntityManager()
	err = mgr.ParseSQLSchema(pgQuery)
	if err != nil {
		return nil, errors.New("failed to parse SQL data: " + err.Error())
	}

	// build dependency graph
	graph := parse.BuildDependencyGraph(mgr)
	result, err := graph.TopologicalSort()
	if err != nil {
		return nil, errors.New("failed to get table relationships: " + err.Error())
	}

	return &Builder{
		sortedTables: result,
	}, nil
}

func (b *Builder) GenerateTableSchemas() ([]GolangFile, error) {
	tableSchemas := utils.Map(b.sortedTables, func(table parse.Table) TableSchema {
		// get the constraint columns
		constraintColumnNames := utils.Reduce(table.Constraints, func(result []string, constraint parse.TableConstraint) []string {
			if constraint.Type == parse.ConstraintInfoTypeForeignKey {
				info := constraint.Constraint.(*parse.ForeignKeyConstraintInfo)
				return append(result, info.TableColumnName)
			}
			return result
		}, []string{})

		// get the dependent tables and columns
		dependencyTables := utils.Reduce(table.Constraints, func(result map[string][]string, constraint parse.TableConstraint) map[string][]string {
			if constraint.Type == parse.ConstraintInfoTypeForeignKey {
				info := constraint.Constraint.(*parse.ForeignKeyConstraintInfo)
				dependencyName := info.ForeignKeyTableName
				colName := info.TableColumnName
				result[dependencyName] = append(result[dependencyName], colName)
				return result
			}
			return result
		}, make(map[string][]string))

		allColumns := utils.Map(table.Columns, func(column parse.Column) TableSchemaColumn {
			goType := checkGoType(column)
			return TableSchemaColumn{
				Name:   column.Name,
				GoType: goType,
			}
		})

		inputColumns := utils.Reduce(table.Columns, func(result []TableSchemaColumn, column parse.Column) []TableSchemaColumn {
			if slices.Contains(constraintColumnNames, column.Name) {
				return result
			}

			// create the input type
			goType := checkGoType(column)
			return append(result, TableSchemaColumn{
				Name:   column.Name,
				GoType: goType,
			})
		}, []TableSchemaColumn{})

		inputToOutputMap := make(map[string]OutputMapData)
		for _, column := range allColumns {
			recordColName := strcase.ToCamel(column.Name)

			added := false
			for _, x := range table.Constraints {
				if x.Type == parse.ConstraintInfoTypeForeignKey {
					info := x.Constraint.(*parse.ForeignKeyConstraintInfo)
					if info.TableColumnName == column.Name {
						fkTableName := info.ForeignKeyTableName
						fkColumnName := info.ForeignKeyColumnName
						inputToOutputMap[recordColName] = OutputMapData{
							ObjectName: strcase.ToCamel(fkTableName) + "Model",
							FieldName:  strcase.ToCamel(fkColumnName),
						}
						added = true
					}
				}
			}

			if !added {
				inputToOutputMap[recordColName] = OutputMapData{
					ObjectName: "input",
					FieldName:  recordColName,
				}
			}
		}

		tableSchema := TableSchema{
			TableName:        table.Name,
			TableColumns:     allColumns,
			InputColumns:     inputColumns,
			DependencyTables: dependencyTables,
			InputToOutputMap: inputToOutputMap,
		}
		return FormatTableSchema(tableSchema)
	})

	// generate the table record files
	files := make([]GolangFile, 0)
	for _, tableSchema := range tableSchemas {
		fileContents, err := GenerateGoFileFromTableSchema(tableSchema)
		if err != nil {
			return nil, err
		}
		files = append(files, GolangFile{
			Filename: strcase.ToSnake(tableSchema.TableName) + ".go",
			Contents: fileContents,
		})
	}

	// generate the seed script
	seedScriptContents, err := GenerateSeedScriptFromTableSchemas(tableSchemas)
	if err != nil {
		return nil, err
	}
	files = append(files, GolangFile{
		Filename: "seed.go",
		Contents: seedScriptContents,
	})

	return files, nil
}

func FormatTableSchema(tableSchema TableSchema) TableSchema {
	newDependencyTables := make(map[string][]string)
	for key, value := range tableSchema.DependencyTables {
		newDependencyTables[strcase.ToCamel(key)] = utils.Map(value, func(column string) string {
			return strcase.ToCamel(column)
		})
	}

	return TableSchema{
		OriginalTableName: tableSchema.TableName,
		TableName:         strcase.ToCamel(tableSchema.TableName),
		OriginalColumnNames: utils.Map(tableSchema.TableColumns, func(column TableSchemaColumn) string {
			return column.Name
		}),
		TableColumns: utils.Map(tableSchema.TableColumns, func(column TableSchemaColumn) TableSchemaColumn {
			return TableSchemaColumn{
				Name:   strcase.ToCamel(column.Name),
				GoType: column.GoType,
			}
		}),
		InputColumns: utils.Map(tableSchema.InputColumns, func(column TableSchemaColumn) TableSchemaColumn {
			return TableSchemaColumn{
				Name:   strcase.ToCamel(column.Name),
				GoType: column.GoType,
			}
		}),
		DependencyTables: newDependencyTables,
		InputToOutputMap: tableSchema.InputToOutputMap,
	}
}

func checkGoType(column parse.Column) string {
	nullable := false
	for _, cons := range column.Constraints {
		if cons.Type == parse.ConstraintInfoTypeDefault {
			if cons.ExpressionValue == "NULL" {
				nullable = true
			}
		}
		if cons.Type == parse.ConstraintInfoTypeNotNull {
			if nullable {
				panic("column is nullable and has not null constraint")
			}
			nullable = false
		}
	}

	goType := "any"
	switch strings.ToLower(column.DataType) {
	case "text":
		goType = "string"
	case "text[]":
		goType = "[]string"
	case "varchar":
		goType = "string"

	case "boolean":
		goType = "bool"

	case "date":
		goType = "time.Time"
	case "timestamptz":
		goType = "time.Time"

	case "smallint":
		goType = "int16"
	case "int2":
		goType = "int16"
	case "int4":
		goType = "int32"
	case "integer":
		goType = "int32"
	case "serial":
		goType = "int32"
	case "bigserial":
		goType = "int64"
	case "int8":
		goType = "int64"
	case "bigint":
		goType = "int64"
	case "real":
		goType = "float32"
	case "float4":
		goType = "float32"
	case "decimal":
		goType = "float64"
	case "numeric":
		goType = "float64"
	case "double precision":
		goType = "float64"

	case "json":
		goType = "map[string]any"

	default:
		goType = "any"
	}

	if nullable && !strings.HasPrefix(goType, "[]") {
		goType = "*" + goType
	}
	return goType
}
