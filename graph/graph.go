/* Package graph is an adjacency list representation of graph.
   The graph can be best visualised as:
   {
	   N1 -> [E1, E2]
	   N2 -> [E2, E3]
	   .
	   .
	   .
	}
	The methods are not made concurrent safe, callers are responsible to take
    Lock before calling them.
*/
package graph

import (
	"errors"
	"fmt"
	"sync"
)

type Edge struct {
	Target       string
	Weight       float64
	Relationship string
}

//NOTE: Only insert members here which do not render it uncomparable for equality.
type Node struct {
	Name string
}

type Graph struct {
	Lock  sync.RWMutex
	Nodes map[Node][]Edge
}

func New() *Graph {
	gr := new(Graph)
	gr.Nodes = map[Node][]Edge{}
	return gr
}

func NewNode(name string) Node {
	return Node{name}
}

func NewEdge(target string, weight float64, relationship string) Edge {
	return Edge{target, weight, relationship}
}

// MUST be called with lock. Note we dont take lock here so that caller
// can do bulk updates as in bootstrapping thereby boosting performance.
func (g *Graph) InsertNode(name string) error {
	if g == nil || g.Nodes == nil {
		return errors.New("graph is not initialised")
	}
	n := NewNode(name)
	return g.insertNode(n)
}

func (g *Graph) insertNode(n Node) error {
	if _, ok := g.Nodes[n]; ok {
		return fmt.Errorf("attempt to override already existing node: %v", n)
	}
	g.Nodes[n] = []Edge{}
	return nil
}

func (g *Graph) UpdateNode(old, new string) error {
	if g == nil || g.Nodes == nil {
		return errors.New("graph is not initialised")
	}

	nOld := NewNode(old)
	nNew := NewNode(new)
	return g.updateNode(nOld, nNew)
}

func (g *Graph) updateNode(nOld, nNew Node) error {
	if eOld, ok := g.Nodes[nOld]; ok {
		if _, ok := g.Nodes[nNew]; !ok {
			for _, edges := range g.Nodes {
				for _, e := range edges {
					if e.Target == nOld.Name {
						e.Target = nNew.Name
					}
				}
			}
			g.deleteNode(nOld)
			g.Nodes[nNew] = eOld
		} else {
			return fmt.Errorf("Attempt to update existing 'to' node: %v", nNew.Name)
		}
	} else {
		return fmt.Errorf("Attempt to update non-existing 'from' node: %v", nOld.Name)
	}
	return nil
}

func (g *Graph) DeleteNode(name string) error {
	if g == nil || g.Nodes == nil {
		return errors.New("Graph is not initialised!")
	}
	n := NewNode(name)
	return g.deleteNode(n)
}

func (g *Graph) deleteNode(n Node) error {
	for n1, edges := range g.Nodes {
		delIndex := -1
		for i, edge := range edges {
			if edge.Target == n.Name {
				delIndex = i
			}
		}
		if delIndex >= 0 {
			//O(1) slice element deletion
			edges[delIndex] = edges[len(edges)-1]
			edges = edges[:len(edges)-1]
			g.Nodes[n1] = edges //update map
		}
	}
	delete(g.Nodes, n)
	return nil
}

func (g *Graph) InsertEdge(source, target string, weight float64, relationship string) error {
	if g == nil || g.Nodes == nil {
		return errors.New("graph is not initialised")
	}

	e := NewEdge(target, weight, relationship)
	nSource := NewNode(source)
	nTarget := NewNode(target)
	if !g.NodeExists(nSource) || !g.NodeExists(nTarget) {
		return fmt.Errorf("attempt to insert Edge between non existing Node %v OR %v", source, target)
	}

	if g.EdgeExists(nSource, nTarget) {
		return errors.New(fmt.Sprintf("attempt to override Edge between Node %v AND %v: ", source, target))
	}

	return g.insertEdge(nSource, e)
}

func (g *Graph) insertEdge(n Node, e Edge) error {
	g.Nodes[n] = append(g.Nodes[n], e)
	return nil
}

func (g *Graph) UpdateEdge(source, target string, weight float64, relationship string) error {
	if g == nil || g.Nodes == nil {
		return errors.New("Graph is not initialised!")
	}

	e := NewEdge(target, weight, relationship)
	nSource := NewNode(source)
	nTarget := NewNode(target)
	if !g.NodeExists(nSource) || !g.NodeExists(nTarget) {
		return errors.New(fmt.Sprintf("Attempt to insert Edge between non existing Node %v OR %v", source, target))
	}

	if !g.EdgeExists(nSource, nTarget) {
		return g.insertEdge(nSource, e)
	}

	return g.updateEdge(nSource, e)
}

func (g *Graph) updateEdge(n Node, e Edge) error {
	edges := []Edge{}
	for _, e1 := range g.Nodes[n] {
		if e1.Target == e.Target {
			e1.Weight = e.Weight
			e1.Relationship = e.Relationship
		}
		edges = append(edges, e1)
	}
	g.Nodes[n] = edges
	return nil
}

func (g *Graph) DeleteEdge(source, target string) error {
	if g == nil || g.Nodes == nil {
		return errors.New("Graph is not initialised!")
	}

	e := NewEdge(target, 0.0, "")
	nSource := NewNode(source)
	nTarget := NewNode(target)
	// Should we return error OR is it okay???
	if !g.NodeExists(nSource) || !g.NodeExists(nTarget) {
		return nil
	}

	return g.deleteEdge(nSource, e)
}

func (g *Graph) deleteEdge(n Node, e Edge) error {
	edges := g.Nodes[n]
	delIndex := -1
	for i, edge := range g.Nodes[n] {
		if edge.Target == e.Target {
			delIndex = i
		}
	}
	if delIndex >= 0 {
		//O(1) slice element deletion
		edges[delIndex] = edges[len(g.Nodes[n])-1]
		edges = edges[:len(edges)-1]
		g.Nodes[n] = edges //update map
	}
	return nil
}

//// INTERNAL FUNCS

func (g *Graph) NodeExists(n Node) bool {
	_, ok := g.Nodes[n]
	return ok
}

func (g *Graph) EdgeExists(s, t Node) bool {
	Edges, ok := g.Nodes[s]
	if !ok {
		return false
	}
	_, ok = g.Nodes[t]
	if !ok {
		return false
	}
	for _, e := range Edges {
		if e.Target == t.Name {
			return true
		}
	}
	return false
}
