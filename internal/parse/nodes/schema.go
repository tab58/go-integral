package nodes

import (
	"fmt"

	pg_query "github.com/pganalyze/pg_query_go/v6"
)

type PostgreSQLSchema struct {
	Tables map[string]Table
}

func NewPostgreSQLSchema(sqlSchema string) (*PostgreSQLSchema, error) {
	// parse the SQL schema text
	pgResult, err := pg_query.Parse(sqlSchema)
	if err != nil {
		return nil, err
	}

	// parse the result and get the table relationships
	tables := make(map[string]Table)
	for _, rawStmt := range pgResult.Stmts {
		stmt := rawStmt.GetStmt()
		switch n := stmt.Node.(type) {
		case *pg_query.Node_CreateStmt:
			table, err := ParsePGTableCreateStatement(n)
			if err != nil {
				return nil, err
			}
			tables[table.Name] = table
		case *pg_query.Node_IndexStmt:
			fmt.Printf("skipping index stmt...\n")
		default:
			return nil, fmt.Errorf("unknown node type: %+v", n)

		}
	}

	return &PostgreSQLSchema{Tables: tables}, nil
}
