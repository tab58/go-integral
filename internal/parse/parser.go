package parse

import (
	"fmt"

	pg_query "github.com/pganalyze/pg_query_go/v6"
)

type EntityManager struct {
	Tables map[string]Table
}

func NewEntityManager() *EntityManager {
	return &EntityManager{
		Tables: make(map[string]Table),
	}
}

func (m *EntityManager) ParseSQLSchema(result *pg_query.ParseResult) error {
	for _, rawStmt := range result.Stmts {
		stmt := rawStmt.GetStmt()
		err := m.parseNode(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *EntityManager) parseNode(node *pg_query.Node) error {
	switch n := node.Node.(type) {
	case *pg_query.Node_CreateStmt:
		return m.parseCreateTable(n)
	case *pg_query.Node_IndexStmt:
		return m.parseIndex(n)
	default:
		return fmt.Errorf("unknown node type: %+v", n)
	}
}
