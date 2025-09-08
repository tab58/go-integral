package parse

import (
	"fmt"

	pg_query "github.com/pganalyze/pg_query_go/v6"
)

func (m *EntityManager) parseIndex(node *pg_query.Node_IndexStmt) error {
	return nil
}

func (m *EntityManager) parseCreateTable(node *pg_query.Node_CreateStmt) error {
	stmt := node.CreateStmt

	// iterate through table elements
	var columns []Column
	var tableConstraints []TableConstraint
	for _, tableElt := range stmt.TableElts {
		switch t := tableElt.Node.(type) {
		case *pg_query.Node_ColumnDef:
			col, err := parseColumnNode(t)
			if err != nil {
				return err
			}
			columns = append(columns, col)
		case *pg_query.Node_Constraint:
			parsed, err := parseTableConstraintNode(t)
			if err != nil {
				return err
			}
			tableConstraints = append(tableConstraints, parsed...)
		}
	}

	m.Tables[stmt.Relation.Relname] = Table{
		Name:        stmt.Relation.Relname,
		Columns:     columns,
		Constraints: tableConstraints,
	}
	return nil
}

func parseTableConstraintNode(node *pg_query.Node_Constraint) ([]TableConstraint, error) {
	constraint := node.Constraint

	switch constraint.Contype {
	case pg_query.ConstrType_CONSTR_FOREIGN:
		tableConstraints := make([]TableConstraint, 0)
		fkAttrs := constraint.FkAttrs
		pkAttrs := constraint.PkAttrs
		if len(fkAttrs) != len(pkAttrs) {
			return nil, fmt.Errorf("foreign key and primary key attributes must have the same length")
		}
		pkTableName := constraint.Pktable.Relname
		for i, fkAttr := range constraint.FkAttrs {
			pkAttr := constraint.PkAttrs[i]
			fkColumnName := fkAttr.Node.(*pg_query.Node_String_).String_.Sval
			pkColumnName := pkAttr.Node.(*pg_query.Node_String_).String_.Sval
			tableConstraints = append(tableConstraints, TableConstraint{
				Type: ConstraintInfoTypeForeignKey,
				Constraint: &ForeignKeyConstraintInfo{
					TableColumnName:      fkColumnName,
					ForeignKeyTableName:  pkTableName,
					ForeignKeyColumnName: pkColumnName,
				},
			})
		}
		return tableConstraints, nil
	case pg_query.ConstrType_CONSTR_PRIMARY:
		columnNames := make([]string, 0)
		for _, key := range constraint.Keys {
			columnNames = append(columnNames, key.Node.(*pg_query.Node_String_).String_.Sval)
		}
		return []TableConstraint{
			{
				Type: ConstraintInfoTypePrimaryKey,
				Constraint: &PrimaryKeyConstraintInfo{
					ColumnNames: columnNames,
				},
			},
		}, nil
	}
	return nil, nil
}

func parseColumnNode(colDef *pg_query.Node_ColumnDef) (Column, error) {
	colName := colDef.ColumnDef.Colname
	colTypeString := getColumnType(colDef)
	colConstraints, err := getColumnConstraints(colDef)
	if err != nil {
		return Column{}, err
	}
	return Column{
		Name:        colName,
		DataType:    colTypeString,
		Constraints: colConstraints,
	}, nil
}

func getColumnType(colDef *pg_query.Node_ColumnDef) string {
	def := colDef.ColumnDef
	colType := def.TypeName

	var colTypeString string = "unknown"

	colTypeNames := colType.Names
	for _, colTypeName := range colTypeNames {
		switch t := colTypeName.Node.(type) {
		case *pg_query.Node_TypeName:
			colTypeString = t.TypeName.String()
		case *pg_query.Node_AConst:
			colTypeString = t.AConst.String()
		case *pg_query.Node_String_:
			colTypeString = t.String_.Sval
		}
	}

	consts, err := getColumnConstraints(colDef)
	if err != nil {
		return colTypeString
	}

	for _, cons := range consts {
		if cons.Type == ConstraintInfoTypeDefault {
			if cons.ExpressionValue == "{}" {
				colTypeString = colTypeString + "[]"
			}
		}
	}

	return colTypeString
}

// func

func getColumnConstraints(colDef *pg_query.Node_ColumnDef) ([]ColumnConstraint, error) {
	var colConstraints []ColumnConstraint
	colDefConstraints := colDef.ColumnDef.Constraints
	for _, cons := range colDefConstraints {
		if node := cons.Node.(*pg_query.Node_Constraint); node != nil {
			constraint := node.Constraint

			constraintExprValue := ""
			constraintType := ConstraintInfoTypeUnknown
			switch typ := constraint.Contype.String(); typ {
			case "CONSTR_PRIMARY":
				constraintType = ConstraintInfoTypePrimaryKey
			case "CONSTR_UNIQUE":
				constraintType = ConstraintInfoTypeUnique
			case "CONSTR_DEFAULT":
				exprNode := constraint.RawExpr.Node
				if e, ok := exprNode.(*pg_query.Node_AConst); ok {
					if e.AConst.Isnull {
						constraintExprValue = "NULL"
					} else {
						value := e.AConst.GetSval().GetSval()
						constraintExprValue = value
					}
				}
				constraintType = ConstraintInfoTypeDefault
			default:
				constraintType = ConstraintInfoType(typ)
			}

			colConstraints = append(colConstraints, ColumnConstraint{
				Type:            constraintType,
				ExpressionValue: constraintExprValue,
			})
		} else {
			return nil, fmt.Errorf("unknown constraint type: %+v", cons.Node)
		}
	}
	return colConstraints, nil
}
