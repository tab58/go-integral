package nodes

import (
	"fmt"
	"log/slog"

	pg_query "github.com/pganalyze/pg_query_go/v6"
)

type TableConstraint struct {
	Type       ConstraintInfoType `json:"type"`
	Constraint ConstraintInfo     `json:"constraint"`
}

type Table struct {
	Name        string            `json:"name"`
	PrimaryKey  []string          `json:"pk"`
	Columns     []Column          `json:"columns"`
	Constraints []TableConstraint `json:"table_constraints"`
}

func ParsePGTableCreateStatement(tableNode *pg_query.Node_CreateStmt) (Table, error) {
	stmt := tableNode.CreateStmt

	tableName := stmt.Relation.Relname
	var columns []Column
	var tableConstraints []TableConstraint

	// iterate through the statements and get information
	for _, tableElt := range stmt.TableElts {
		switch t := tableElt.Node.(type) {
		case *pg_query.Node_ColumnDef:
			col, err := ParsePGColumnDefinition(t)
			if err != nil {
				return Table{}, err
			}
			columns = append(columns, col)
		case *pg_query.Node_Constraint:
			constraint, err := ParsePGTableConstraints(t)
			if err != nil {
				return Table{}, err
			}
			tableConstraints = append(tableConstraints, constraint...)
		default:
			slog.Info("found unknown table element type")
		}
	}

	// find out what the primary key is
	pkColumns := make([]string, 0)
	for _, tableCon := range tableConstraints {
		if tableCon.Type == ConstraintInfoTypePrimaryKey {
			pk, ok := tableCon.Constraint.(*PrimaryKeyConstraintInfo)
			if !ok {
				return Table{}, fmt.Errorf("constraint cannot be converted to a primary key constraint")
			}
			pkColumns = append(pkColumns, pk.ColumnNames...)
		}
	}
	if len(pkColumns) == 0 {
		for _, col := range columns {
			for _, colCon := range col.Constraints {
				if colCon.Type == ConstraintInfoTypePrimaryKey {
					pkColumns = append(pkColumns, col.Name)
				}
			}
		}
	}

	return Table{
		Name:        tableName,
		PrimaryKey:  pkColumns,
		Columns:     columns,
		Constraints: tableConstraints,
	}, nil
}
