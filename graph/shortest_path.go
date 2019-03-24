package graph

import (
	"container/heap"

	"github.com/alokic/gopkg/mathutils"
)

type shortestPathTerm struct {
	Node Node
	Cost float64
}

type HeapNode struct {
	N    Node
	Dist float64
	Disp uint32
}
type HeapNodes []HeapNode

func (h *HeapNodes) Len() int           { return len(*h) }
func (h *HeapNodes) Less(i, j int) bool { return (*h)[i].Dist < (*h)[j].Dist }
func (h *HeapNodes) Swap(i, j int)      { (*h)[i], (*h)[j] = (*h)[j], (*h)[i] }

func (h *HeapNodes) Push(x interface{}) {
	*h = append(*h, x.(HeapNode))
}

func (h *HeapNodes) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// Only for non-negative edge weights
func (g *Graph) Dijikstra(n Node, maxCost float64, maxHops uint32) ([]*shortestPathTerm, error) {

	h := &HeapNodes{}
	dist := make(map[Node]float64, 1000)
	disp := make(map[Node]uint32, 1000)
	visited := make(map[Node]float64, 1000)

	heap.Init(h)
	heap.Push(h, HeapNode{n, 0.0, 0}) //push source

	for h.Len() > 0 {
		curr := heap.Pop(h).(HeapNode) //PERF
		// This might be a duplicate node in the heap, so ignore if visited
		if _, ok := visited[curr.N]; !ok {
			visited[curr.N] = curr.Dist
			if curr.Disp < maxHops {
				//Iterate over neighbour nodes
				for _, edge := range g.Nodes[curr.N] {
					neighbourNode := NewNode(edge.Target)
					// We are only interested in non visited nodes
					if _, ok := visited[neighbourNode]; !ok {
						neighbourDist := curr.Dist + edge.Weight
						neighbourDisp := curr.Disp + 1
						// If the node is in heap
						if val, ok := dist[neighbourNode]; ok {
							if val == curr.Dist+edge.Weight {
								//choose min hops
								neighbourDisp = uint32(mathutils.MinInt(int(neighbourDisp), int(disp[neighbourNode])))
							} else if val < curr.Dist+edge.Weight {
								neighbourDist = val
								neighbourDisp = disp[neighbourNode]
							}
						}
						// If this node is affordable
						if neighbourDist <= maxCost {
							//OK to pass duplicates even if priority is same
							heap.Push(h, HeapNode{neighbourNode, neighbourDist, neighbourDisp}) //PERF
							dist[neighbourNode] = neighbourDist
							disp[neighbourNode] = neighbourDisp
						}
					}
				}
			}
		}
	}

	result := make([]*shortestPathTerm, len(visited))
	i := 0
	for node, cost := range visited {
		result[i] = &shortestPathTerm{node, cost}
		i++
	}
	return result, nil
}
