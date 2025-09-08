package graph

import (
	"errors"

	"github.com/segmentio/ksuid"
)

type Node[T any] struct {
	ID    string `json:"id"`
	Value T      `json:"value"`
}

type DirectedEdge[T any, E any] struct {
	From  *Node[T]
	To    *Node[T]
	Value E
}

type DirectedGraph[T any, E any] struct {
	Nodes []*Node[T]
	Edges []*DirectedEdge[T, E]
}

func NewDirectedGraph[T any, E any]() *DirectedGraph[T, E] {
	return &DirectedGraph[T, E]{
		Nodes: make([]*Node[T], 0),
		Edges: make([]*DirectedEdge[T, E], 0),
	}
}

func (g *DirectedGraph[T, E]) AddNode(nodeValue T) *Node[T] {
	node := &Node[T]{
		ID:    ksuid.New().String(),
		Value: nodeValue,
	}
	g.Nodes = append(g.Nodes, node)
	return node
}

func (g *DirectedGraph[T, E]) AddEdge(from *Node[T], to *Node[T], edgeValue E) {
	g.Edges = append(g.Edges, &DirectedEdge[T, E]{
		From:  from,
		To:    to,
		Value: edgeValue,
	})
}

type topoNode[T any] struct {
	Node           *Node[T]
	InDegree       int
	ConnectedNodes []*topoNode[T]
}

func (g *DirectedGraph[T, E]) TopologicalSort() ([]T, error) {
	// build topoNodes
	topoNodes := make(map[string]*topoNode[T])
	for _, n := range g.Nodes {
		topoNodes[n.ID] = &topoNode[T]{
			Node:           n,
			InDegree:       0,
			ConnectedNodes: make([]*topoNode[T], 0),
		}
	}

	// calculate in-degree nodes
	for _, edge := range g.Edges {
		fromNode := topoNodes[edge.From.ID]
		toNode := topoNodes[edge.To.ID]

		connNodes := fromNode.ConnectedNodes
		connNodes = append(connNodes, toNode)
		fromNode.ConnectedNodes = connNodes
		toNode.InDegree++
	}

	// enqueue nodes with in-degree 0
	q := Queue[topoNode[T]]{}
	for _, node := range topoNodes {
		if node.InDegree == 0 {
			q.Enqueue(*node)
		}
	}

	// iterate through nodes
	result := make([]*Node[T], 0)
	for !q.IsEmpty() {
		n := q.Dequeue()
		result = append(result, n.Node)

		if len(n.ConnectedNodes) != 0 {
			for _, connectedNode := range n.ConnectedNodes {
				connectedNode.InDegree--
				if connectedNode.InDegree == 0 {
					q.Enqueue(*connectedNode)
				}
			}
		}
	}

	if len(result) != len(topoNodes) {
		return nil, errors.New("graph has a cycle")
	}

	resultValues := make([]T, 0)
	for _, n := range result {
		resultValues = append(resultValues, n.Value)
	}
	return resultValues, nil
}
