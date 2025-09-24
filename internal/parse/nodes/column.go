package nodes

import (
	"fmt"

	pg_query "github.com/pganalyze/pg_query_go/v6"
)

type ColumnConstraint struct {
	Type            ConstraintInfoType `json:"type"`
	ExpressionValue string             `json:"expression_value"`
}

type Column struct {
	Name        string             `json:"name"`
	DataType    string             `json:"data_type"`
	Constraints []ColumnConstraint `json:"constraints"`
}

func ParsePGColumnDefinition(colDef *pg_query.Node_ColumnDef) (Column, error) {
	colName := colDef.ColumnDef.Colname
	colTypeString := ParsePGColumnDataType(colDef)
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

func ParsePGColumnDataType(colDef *pg_query.Node_ColumnDef) string {
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
			// TODO: this is a HACK to detect arrays because pg_query_go doesn't recognize them
			// This should be removed once pg_query_go recognizes arrays
			if cons.ExpressionValue == "{}" {
				colTypeString = colTypeString + "[]"
			}
		}
	}

	return colTypeString
}

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
