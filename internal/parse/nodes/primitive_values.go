package nodes

import pg_query "github.com/pganalyze/pg_query_go/v6"

func ParsePGAConst(e *pg_query.Node_AConst) string {
	if e.AConst.Isnull {
		return "NULL"
	}
	value := e.AConst.GetSval().GetSval()
	return value
}
