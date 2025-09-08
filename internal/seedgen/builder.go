package seedgen

import (
	"errors"
	"go-integral/internal/parse"
	"go-integral/internal/utils"
	"slices"

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
	// res, err := json.Marshal(b.sortedTables)
	// if err != nil {
	// 	return nil, errors.New("failed to marshal sorted tables: " + err.Error())
	// }
	// fmt.Printf("%+v\n", string(res))

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
			return TableSchemaColumn{
				Name:   column.Name,
				GoType: checkGoType(column),
			}
		})

		inputColumns := utils.Reduce(table.Columns, func(result []TableSchemaColumn, column parse.Column) []TableSchemaColumn {
			if slices.Contains(constraintColumnNames, column.Name) {
				return result
			}

			// create the input type
			return append(result, TableSchemaColumn{
				Name:   column.Name,
				GoType: checkGoType(column),
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
	isArray := false
	for _, cons := range column.Constraints {
		if cons.Type == parse.ConstraintInfoTypeDefault {
			if cons.ExpressionValue == "NULL" {
				nullable = true
			}
			if cons.ExpressionValue == "{}" { // this is a HACK to detect arrays because pq_query_go doesn't recognize them
				isArray = true
			}
		}
		if cons.Type == parse.ConstraintInfoTypeNotNull {
			if nullable {
				panic("column is nullable and has not null constraint")
			}
			nullable = false
		}
	}

	goType := "string" // default to string

	switch column.DataType {
	case "text":
		goType = "string"
	case "integer":
		goType = "int"
	case "boolean":
		goType = "bool"
	case "timestamptz":
		goType = "time.Time"
	}

	if nullable {
		goType = "*" + goType
	}
	if isArray {
		goType = "[]" + goType
	}
	return goType
}
