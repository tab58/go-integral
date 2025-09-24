package parse

import (
	"fmt"
	"go-integral/internal/graph"
	"go-integral/internal/parse/nodes"
)

type TableDependencyNode struct {
	Node        *graph.Node[nodes.Table]
	TableName   string
	TableColumn string
}

type TableDependency struct {
	FromNode TableDependencyNode
	ToNode   TableDependencyNode
}

func BuildSQLTableGraph(sqlSchema string) (*graph.DirectedGraph[nodes.Table, TableDependency], error) {
	schema, err := nodes.NewPostgreSQLSchema(sqlSchema)
	if err != nil {
		return nil, fmt.Errorf("unable to parse SQL schema: %w", err)
	}

	schemaGraph := graph.NewDirectedGraph[nodes.Table, TableDependency]()

	// add nodes in the graph
	tableNodes := make(map[string]*graph.Node[nodes.Table])
	for tableName, table := range schema.Tables {
		tableNodes[tableName] = schemaGraph.AddNode(table)
	}
	for tableName, table := range schema.Tables {
		for _, constraint := range table.Constraints {
			if constraint.Type == nodes.ConstraintInfoTypeForeignKey {
				dependency, err := buildGraphTableEdge(constraint, tableName, tableNodes)
				if err != nil {
					return nil, fmt.Errorf("could not build table graph edge: %w", err)
				}
				schemaGraph.AddEdge(dependency.FromNode.Node, dependency.ToNode.Node, dependency)

			}
		}
	}

	return schemaGraph, nil
}

func buildGraphTableEdge(fk nodes.TableConstraint, tableName string, tableNodes map[string]*graph.Node[nodes.Table]) (TableDependency, error) {
	info, ok := fk.Constraint.(*nodes.ForeignKeyConstraintInfo)
	if !ok {
		return TableDependency{}, fmt.Errorf("table constraint cannot be converted to a foreign key constraint")
	}

	fromNode := tableNodes[info.ForeignKeyTableName]
	toNode := tableNodes[tableName]

	fromTable := info.ForeignKeyTableName
	fromColumn := info.ForeignKeyColumnName
	toTable := tableName
	toColumn := info.TableColumnName

	return TableDependency{
		FromNode: TableDependencyNode{
			Node:        fromNode,
			TableName:   fromTable,
			TableColumn: fromColumn,
		},
		ToNode: TableDependencyNode{
			Node:        toNode,
			TableName:   toTable,
			TableColumn: toColumn,
		},
	}, nil
}
