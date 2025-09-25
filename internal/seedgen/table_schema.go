package seedgen

import (
	"fmt"
	"go-integral/internal/parse/nodes"
	"go-integral/internal/utils"
	"slices"

	"github.com/iancoleman/strcase"
)

func generateTableSchema(table nodes.Table) (TableSchema, error) {
	// get the constraint columns
	constraintColumnNames, err := generateConstraintColumnNames(table.Constraints)
	if err != nil {
		return TableSchema{}, fmt.Errorf("unable to generate column constraint names: %w", err)
	}

	// get the dependent tables and columns
	dependencyTables, err := getDependentTables(table.Constraints)
	if err != nil {
		return TableSchema{}, fmt.Errorf("unable to generate dependent tables: %w", err)
	}

	// convert all the columns to schema columns
	allColumns := utils.Map(table.Columns, convertToSchemaColumn)

	// filter on the constraints to get the input columns
	inputColumns := utils.Filter(allColumns, func(column RawTableSchemaColumn) bool {
		return !slices.Contains(constraintColumnNames, column.Name)
	})

	// map the Golang input names to the SQL output names
	inputToOutputMap, err := createInputOutputMap(allColumns, table.Constraints)
	if err != nil {
		return TableSchema{}, fmt.Errorf("unable to get input-output map: %w", err)
	}

	tableSchema := RawTableSchema{
		TableName:        table.Name,
		TablePrimaryKey:  table.PrimaryKey,
		TableColumns:     allColumns,
		InputColumns:     inputColumns,
		DependencyTables: dependencyTables,
		InputToOutputMap: inputToOutputMap,
	}
	return refineTableSchema(tableSchema), nil
}

// refineTableSchema "massages" the format of the table schema to make it more Golang-friendly
func refineTableSchema(tableSchema RawTableSchema) TableSchema {
	refinedTableSchema := TableSchema{
		TableName: SQLGolangStringValue{
			SQL:    tableSchema.TableName,
			Golang: strcase.ToCamel(tableSchema.TableName),
		},
		SQLTablePrimaryKey: utils.Map(tableSchema.TablePrimaryKey, func(column string) SQLGolangStringValue {
			return SQLGolangStringValue{
				SQL:    column,
				Golang: strcase.ToCamel(column),
			}
		}),
		TableColumns: utils.Map(tableSchema.TableColumns, func(column RawTableSchemaColumn) TableSchemaColumn {
			return TableSchemaColumn{
				Name: SQLGolangStringValue{
					SQL:    column.Name,
					Golang: strcase.ToCamel(column.Name),
				},
				GoType: column.GoType,
			}
		}),
		RecordInputColumns: utils.Map(tableSchema.InputColumns, func(column RawTableSchemaColumn) TableSchemaColumn {
			return TableSchemaColumn{
				Name: SQLGolangStringValue{
					SQL:    column.Name,
					Golang: strcase.ToCamel(column.Name),
				},
				GoType: column.GoType,
			}
		}),
		DependencyTables: refineDependencyTables(tableSchema.DependencyTables),
		InputToOutputMap: tableSchema.InputToOutputMap,
	}
	return refinedTableSchema
}

func refineDependencyTables(dependencyTables map[string][]string) []DependencyTable {
	newDependencyTables := make([]DependencyTable, 0)
	for key, value := range dependencyTables {
		depTableKey := strcase.ToCamel(key)
		newDependencyTables = append(newDependencyTables, DependencyTable{
			GolangTableName: depTableKey,
			InputRecordName: strcase.ToLowerCamel(key),
			ColumnNames: utils.Map(value, func(column string) string {
				return strcase.ToCamel(column)
			}),
		})
	}
	return newDependencyTables
}

func createInputOutputMap(columns []RawTableSchemaColumn, tableConstraints []nodes.TableConstraint) (map[string]OutputMapData, error) {
	inputToOutputMap := make(map[string]OutputMapData)
	for _, column := range columns {
		recordColName := strcase.ToCamel(column.Name)

		added := false
		for _, x := range tableConstraints {
			if x.Type == nodes.ConstraintInfoTypeForeignKey {
				info, ok := x.Constraint.(*nodes.ForeignKeyConstraintInfo)
				if !ok {
					return nil, fmt.Errorf("cannot convert constraint to foreign key constraint")
				}
				if info.TableColumnName == column.Name {
					fkTableName := info.ForeignKeyTableName
					fkColumnName := info.ForeignKeyColumnName
					inputToOutputMap[recordColName] = OutputMapData{
						ObjectName: strcase.ToLowerCamel(fkTableName) + "Model",
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
	return inputToOutputMap, nil
}

func convertToSchemaColumn(column nodes.Column) RawTableSchemaColumn {
	goType := checkGolangDataType(column)
	return RawTableSchemaColumn{
		Name:   column.Name,
		GoType: goType,
	}
}

func getDependentTables(constraints []nodes.TableConstraint) (map[string][]string, error) {
	return utils.ReduceErr(constraints, func(result map[string][]string, constraint nodes.TableConstraint) (map[string][]string, error) {
		if constraint.Type == nodes.ConstraintInfoTypeForeignKey {
			info, ok := constraint.Constraint.(*nodes.ForeignKeyConstraintInfo)
			if !ok {
				return nil, fmt.Errorf("unable to convert constraint to foreign key constraint")
			}

			dependencyName := info.ForeignKeyTableName
			colName := info.TableColumnName
			result[dependencyName] = append(result[dependencyName], colName)
		}
		return result, nil
	}, make(map[string][]string))
}

func generateConstraintColumnNames(constraints []nodes.TableConstraint) ([]string, error) {
	return utils.ReduceErr(constraints, func(result []string, constraint nodes.TableConstraint) ([]string, error) {
		if constraint.Type == nodes.ConstraintInfoTypeForeignKey {
			info, ok := constraint.Constraint.(*nodes.ForeignKeyConstraintInfo)
			if !ok {
				return nil, fmt.Errorf("unable to convert constraint to foreign key constraint")
			}
			return append(result, info.TableColumnName), nil
		}
		return result, nil
	}, []string{})
}
