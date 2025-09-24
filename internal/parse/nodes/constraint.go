package nodes

import (
	"fmt"
	"log/slog"

	pg_query "github.com/pganalyze/pg_query_go/v6"
)

func ParsePGColumnConstraint(constraintNode *pg_query.Node_Constraint) (ColumnConstraint, error) {
	constraint := constraintNode.Constraint

	constraintExprValue := ""
	constraintType := ConstraintInfoTypeUnknown
	switch typ := constraint.Contype; typ {
	case pg_query.ConstrType_CONSTR_PRIMARY:
		constraintType = ConstraintInfoTypePrimaryKey
	case pg_query.ConstrType_CONSTR_UNIQUE:
		constraintType = ConstraintInfoTypeUnique
	case pg_query.ConstrType_CONSTR_DEFAULT:
		exprNode := constraint.RawExpr.Node
		if e, ok := exprNode.(*pg_query.Node_AConst); ok {
			constraintExprValue = ParsePGAConst(e)
		}
		constraintType = ConstraintInfoTypeDefault
	default:
		slog.Info(fmt.Sprintf("adding unknown constraint type %s", typ))
		constraintType = ConstraintInfoType(typ)
	}

	return ColumnConstraint{
		Type:            constraintType,
		ExpressionValue: constraintExprValue,
	}, nil
}

func ParsePGTableConstraints(constraintNode *pg_query.Node_Constraint) ([]TableConstraint, error) {
	constraint := constraintNode.Constraint

	switch constraint.Contype {
	case pg_query.ConstrType_CONSTR_FOREIGN:
		return ParsePGTableForeignKeyConstraints(constraint)
	case pg_query.ConstrType_CONSTR_PRIMARY:
		return ParsePGTablePrimaryKeyConstraints(constraint)
	default:
		slog.Info(fmt.Sprintf("unknown constraint type %s", constraint.Contype))
	}

	return nil, nil
}

func ParsePGTablePrimaryKeyConstraints(constraint *pg_query.Constraint) ([]TableConstraint, error) {
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

func ParsePGTableForeignKeyConstraints(constraint *pg_query.Constraint) ([]TableConstraint, error) {
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
}
