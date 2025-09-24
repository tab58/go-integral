package seedgen

import (
	"errors"
	"go-integral/internal/parse"
	"go-integral/internal/parse/nodes"
	"go-integral/internal/utils"
	"slices"
	"strings"

	"github.com/iancoleman/strcase"
)

type Builder struct {
	sortedTables []nodes.Table
}

func NewFromSQLSchema(sqlSchema string) (*Builder, error) {
	// build dependency graph
	graph, err := parse.BuildSQLTableGraph(sqlSchema)
	if err != nil {
		return nil, errors.New("failed to build table graph: " + err.Error())
	}

	// get tables in record insert order
	result, err := graph.TopologicalSort()
	if err != nil {
		return nil, errors.New("failed to get table relationships: " + err.Error())
	}

	return &Builder{
		sortedTables: result,
	}, nil
}

func (b *Builder) GenerateTableSchemas() ([]GolangFile, error) {
	tableSchemas := utils.Map(b.sortedTables, func(table nodes.Table) TableSchema {
		// get the constraint columns
		constraintColumnNames := utils.Reduce(table.Constraints, func(result []string, constraint nodes.TableConstraint) []string {
			if constraint.Type == nodes.ConstraintInfoTypeForeignKey {
				info := constraint.Constraint.(*nodes.ForeignKeyConstraintInfo)
				return append(result, info.TableColumnName)
			}
			return result
		}, []string{})

		// get the dependent tables and columns
		dependencyTables := utils.Reduce(table.Constraints, func(result map[string][]string, constraint nodes.TableConstraint) map[string][]string {
			if constraint.Type == nodes.ConstraintInfoTypeForeignKey {
				info := constraint.Constraint.(*nodes.ForeignKeyConstraintInfo)
				dependencyName := info.ForeignKeyTableName
				colName := info.TableColumnName
				result[dependencyName] = append(result[dependencyName], colName)
				return result
			}
			return result
		}, make(map[string][]string))

		allColumns := utils.Map(table.Columns, func(column nodes.Column) TableSchemaColumn {
			goType := checkGoType(column)
			return TableSchemaColumn{
				Name:   column.Name,
				GoType: goType,
			}
		})

		inputColumns := utils.Reduce(table.Columns, func(result []TableSchemaColumn, column nodes.Column) []TableSchemaColumn {
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
				if x.Type == nodes.ConstraintInfoTypeForeignKey {
					info := x.Constraint.(*nodes.ForeignKeyConstraintInfo)
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

func checkGoType(column nodes.Column) string {
	nullable := false
	for _, cons := range column.Constraints {
		if cons.Type == nodes.ConstraintInfoTypeDefault {
			if cons.ExpressionValue == "NULL" {
				nullable = true
			}
		}
		if cons.Type == nodes.ConstraintInfoTypeNotNull {
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
