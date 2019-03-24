package graph

import "github.com/alokic/gopkg/queue"

type bfsTerm struct {
	Node  Node
	Level uint
}

func (g *Graph) SubGraph(n Node, level uint) *Graph {
	q := queue.New()
	q.Push(bfsTerm{n, 0})
	visited := make(map[Node]uint, 1000)
	resGraph := New()
	for !q.Empty() {
		curr := q.Pop().(bfsTerm)
		visited[curr.Node] = curr.Level
		resGraph.insertNode(curr.Node)
		if curr.Level < level {
			//Iterate over neighbour nodes
			for _, edge := range g.Nodes[curr.Node] { //PERF
				neighbourNode := NewNode(edge.Target)
				resGraph.insertNode(neighbourNode)
				resGraph.insertEdge(curr.Node, edge)
				// We are only interested in non visited nodes
				if _, ok := visited[neighbourNode]; !ok {
					q.Push(bfsTerm{neighbourNode, curr.Level + 1})
				}
			}
		}
	}

	nodes := make([]Node, len(visited))
	i := 0
	for node := range visited {
		nodes[i] = node
		i++
	}
	return resGraph
}
