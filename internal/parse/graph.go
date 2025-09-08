package parse

import (
	"go-integral/internal/graph"
)

type TableDependency struct {
	FromTable  string
	FromColumn string
	ToTable    string
	ToColumn   string
}

func BuildDependencyGraph(m *EntityManager) *graph.DirectedGraph[Table, TableDependency] {
	tableNodes := make(map[string]*graph.Node[Table])

	graph := graph.NewDirectedGraph[Table, TableDependency]()
	for tableName, table := range m.Tables {
		tableNodes[tableName] = graph.AddNode(table)
	}

	for tableName, table := range m.Tables {
		for _, constraint := range table.Constraints {
			if constraint.Type == ConstraintInfoTypeForeignKey {
				info := constraint.Constraint.(*ForeignKeyConstraintInfo)
				fromNode := tableNodes[info.ForeignKeyTableName]
				toNode := tableNodes[tableName]

				fromTable := info.ForeignKeyTableName
				fromColumn := info.ForeignKeyColumnName
				toTable := tableName
				toColumn := info.TableColumnName

				graph.AddEdge(fromNode, toNode, TableDependency{
					FromTable:  fromTable,
					FromColumn: fromColumn,
					ToTable:    toTable,
					ToColumn:   toColumn,
				})
			}
		}
	}

	return graph
}
