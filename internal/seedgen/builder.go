package seedgen

import (
	"errors"
	"fmt"
	"go-integral/internal/parse"
	"go-integral/internal/parse/nodes"
	"go-integral/internal/utils"
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

func (b *Builder) GenerateTemplateFiles() ([]GolangFile, error) {
	// generate the table schemas
	tableSchemas, err := utils.MapErr(b.sortedTables, generateTableSchema)
	if err != nil {
		return nil, fmt.Errorf("unable to generate table schemas: %w", err)
	}

	// generate the table record files
	files, err := utils.MapErr(tableSchemas, generateGoFileFromTableSchema)
	if err != nil {
		return nil, fmt.Errorf("unable to generate Golang file from table schema: %w", err)
	}

	// generate the seed script
	seedScript, err := generateSeedScriptFromTableSchemas(tableSchemas)
	if err != nil {
		return nil, fmt.Errorf("unable to generate seed script from table schemas: %w", err)
	}

	return append(files, seedScript), nil
}

func generateSeedScriptFromTableSchemas(schemas []TableSchema) (GolangFile, error) {
	contents, err := generateSeedScriptContentsFromTableSchemas(schemas)
	if err != nil {
		return GolangFile{}, fmt.Errorf("unable to generate seed script contents from table schemas: %w", err)
	}
	return GolangFile{
		Filename: "seed.go",
		Contents: contents,
	}, nil
}

func generateGoFileFromTableSchema(schema TableSchema) (GolangFile, error) {
	contents, err := generateFileContentsFromTableSchema(schema)
	if err != nil {
		return GolangFile{}, fmt.Errorf("unable to generate file contents from table schema: %w", err)
	}
	return GolangFile{
		Filename: strcase.ToSnake(schema.TableName) + ".go",
		Contents: contents,
	}, nil
}

func checkGolangDataType(column nodes.Column) string {
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
