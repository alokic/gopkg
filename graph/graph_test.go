package graph

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGraph(t *testing.T) {
	g := New()

	assert.Equal(t, 0, len(g.Nodes), "empty graph")

	//INSERT NODE
	assert.NoError(t, g.InsertNode("alpha"), "insert node")
	assert.NoError(t, g.InsertNode("beta"), "insert node")
	assert.Error(t, g.InsertNode("beta"), "duplicate insert node")
	assert.Equal(t, 2, len(g.Nodes), "2 node graph")
	assert.Equal(t, 0, len(g.Nodes[NewNode("alpha")]), "empty edges list")
	assert.Equal(t, 0, len(g.Nodes[NewNode("beta")]), "empty edges list")
	assert.Equal(t, true, g.NodeExists(NewNode("alpha")), "node exists")
	assert.Equal(t, true, g.NodeExists(NewNode("beta")), "node exists")

	//UPDATE NODE
	assert.Error(t, g.UpdateNode("zeta", "omega"), "update on non-existing from node")
	assert.Error(t, g.UpdateNode("alpha", "beta"), "update on existing to node")
	assert.NoError(t, g.UpdateNode("alpha", "gamma"), "update")
	assert.NoError(t, g.UpdateNode("beta", "zeta"), "update")
	assert.Equal(t, 2, len(g.Nodes), "2 node graph")
	assert.Equal(t, 0, len(g.Nodes[NewNode("alpha")]), "empty edges list")
	assert.Equal(t, 0, len(g.Nodes[NewNode("beta")]), "empty edges list")
	assert.Equal(t, true, g.NodeExists(NewNode("gamma")), "node exists")
	assert.Equal(t, true, g.NodeExists(NewNode("gamma")), "node exists")
	assert.Equal(t, false, g.NodeExists(NewNode("beta")), "node exists")
	assert.Equal(t, false, g.NodeExists(NewNode("alpha")), "node exists")

	//DELETE NODE
	assert.NoError(t, g.DeleteNode("alpha"), "delete non existing node")
	assert.NoError(t, g.DeleteNode("beta"), "delete non existing node")
	assert.Equal(t, 2, len(g.Nodes), "2 node graph")
	assert.NoError(t, g.DeleteNode("zeta"), "delete existing node")
	assert.Equal(t, 1, len(g.Nodes), "1 node graph")
	assert.Equal(t, false, g.NodeExists(NewNode("zeta")), "node not exists")
	assert.Equal(t, true, g.NodeExists(NewNode("gamma")), "node exists")
	assert.NoError(t, g.DeleteNode("gamma"), "delete existing node")
	assert.Equal(t, 0, len(g.Nodes), "empty graph")
	assert.Equal(t, false, g.NodeExists(NewNode("gamma")), "node not exists")

	//INSERT EDGE
	assert.Error(t, g.InsertEdge("alpha", "beta", 0.1, "is"), "non existing source/target node")
	g.insertNode(NewNode("alpha"))
	g.insertNode(NewNode("beta"))
	assert.Equal(t, false, g.EdgeExists(NewNode("alpha"), NewNode("beta")), "edge exists")
	assert.NoError(t, g.InsertEdge("alpha", "beta", 0.1, "is"), "insert edge")
	assert.Error(t, g.InsertEdge("alpha", "beta", 0.1, "is"), "insert duplicate edge")
	assert.Equal(t, true, g.EdgeExists(NewNode("alpha"), NewNode("beta")), "edge exists")
	assert.Equal(t, 1, len(g.Nodes[NewNode("alpha")]), "1 edges list")
	assert.Equal(t, 0, len(g.Nodes[NewNode("beta")]), "0 edges list")

	//UPDATE EDGE
	assert.Error(t, g.UpdateEdge("gamma", "beta", 0.1, "is"), "non existing source/target node")
	assert.NoError(t, g.UpdateEdge("beta", "alpha", 0.1, "is"), "non existing edge becomes insert")
	assert.Equal(t, NewEdge("alpha", 0.1, "is"), g.Nodes[NewNode("beta")][0], "")
	assert.NoError(t, g.UpdateEdge("beta", "alpha", 0.2, "is"), "update existing edge's weight")
	assert.Equal(t, NewEdge("alpha", 0.2, "is"), g.Nodes[NewNode("beta")][0], "")
	assert.NoError(t, g.UpdateEdge("beta", "alpha", 0.2, "has"), "update existing edge's rel")
	assert.Equal(t, NewEdge("alpha", 0.2, "has"), g.Nodes[NewNode("beta")][0], "")

	//DELETE EDGE
	assert.NoError(t, g.InsertNode("gamma"), "")
	assert.NoError(t, g.DeleteEdge("ddd", "alpha"), "non existing node")
	assert.NoError(t, g.DeleteEdge("gamma", "alpha"), "non existing edge")
	assert.Equal(t, true, g.EdgeExists(NewNode("beta"), NewNode("alpha")), "edge exists")
	assert.NoError(t, g.DeleteEdge("beta", "alpha"), "existing edge")
	assert.Equal(t, false, g.EdgeExists(NewNode("beta"), NewNode("alpha")), "no edge exists")
	assert.Equal(t, []Edge{}, g.Nodes[NewNode("beta")], "no edge exists")
	assert.NoError(t, g.DeleteEdge("beta", "alpha"), "OK on non existing edge")
	assert.Equal(t, []Edge{NewEdge("beta", 0.1, "is")}, g.Nodes[NewNode("alpha")], "no edge exists")

}
