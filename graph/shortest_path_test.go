package graph

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ShortestPathTestSuite struct {
	suite.Suite
	g *Graph
}

func (suite *ShortestPathTestSuite) SetupSuite() {
	g := New()
	g.InsertNode("A")
	g.InsertNode("B")
	g.InsertNode("C")
	g.InsertNode("D")
	g.InsertNode("E")
	g.InsertNode("F")
	g.InsertNode("G")
	g.InsertNode("H")
	g.InsertEdge("A", "B", 20, "is")
	g.InsertEdge("A", "D", 80, "is")
	g.InsertEdge("A", "F", 40, "is")
	g.InsertEdge("A", "G", 90, "is")
	g.InsertEdge("B", "F", 10, "is")
	g.InsertEdge("B", "G", 10, "is")
	g.InsertEdge("C", "D", 10, "is")
	g.InsertEdge("C", "F", 50, "is")
	g.InsertEdge("C", "H", 20, "is")
	g.InsertEdge("D", "G", 20, "is")
	g.InsertEdge("E", "B", 5, "is")
	g.InsertEdge("E", "G", 30, "is")
	g.InsertEdge("F", "C", 10, "is")
	g.InsertEdge("F", "D", 40, "is")
	g.InsertEdge("G", "A", 20, "is")
	suite.g = g
}

func (suite *ShortestPathTestSuite) TestDijikstra() {
	tests := map[string]struct {
		source  Node
		maxW    float64
		maxHops uint32
		output  []*shortestPathTerm
	}{
		"shortest path with 0 hops": {
			source:  NewNode("A"),
			maxW:    100,
			maxHops: 0,
			output:  []*shortestPathTerm{&shortestPathTerm{NewNode("A"), 0}},
		},
		"shortest path with 1 hops and no limit on weight": {
			source:  NewNode("A"),
			maxW:    100,
			maxHops: 1,
			output: []*shortestPathTerm{
				&shortestPathTerm{NewNode("A"), 0},
				&shortestPathTerm{NewNode("B"), 20},
				&shortestPathTerm{NewNode("F"), 40},
				&shortestPathTerm{NewNode("D"), 80},
				&shortestPathTerm{NewNode("G"), 90},
			},
		},
		"shortest path with 1 hops but limit on weight-1": {
			source:  NewNode("A"),
			maxW:    10,
			maxHops: 1,
			output: []*shortestPathTerm{
				&shortestPathTerm{NewNode("A"), 0},
			},
		},
		"shortest path with 1 hops but limit on weight-2": {
			source:  NewNode("A"),
			maxW:    20,
			maxHops: 1,
			output: []*shortestPathTerm{
				&shortestPathTerm{NewNode("A"), 0},
				&shortestPathTerm{NewNode("B"), 20},
			},
		},
		"shortest path with 1 hops but limit on weight-3": {
			source:  NewNode("A"),
			maxW:    89,
			maxHops: 1,
			output: []*shortestPathTerm{
				&shortestPathTerm{NewNode("A"), 0},
				&shortestPathTerm{NewNode("B"), 20},
				&shortestPathTerm{NewNode("F"), 40},
				&shortestPathTerm{NewNode("D"), 80},
			},
		},
		"shortest path with 2 hops and no limit on weight": {
			source:  NewNode("A"),
			maxW:    200,
			maxHops: 2,
			output: []*shortestPathTerm{
				&shortestPathTerm{NewNode("A"), 0},
				&shortestPathTerm{NewNode("B"), 20},
				&shortestPathTerm{NewNode("D"), 80},
				&shortestPathTerm{NewNode("F"), 30},
				&shortestPathTerm{NewNode("G"), 30},
			},
		},
		"shortest path with 2 hops and limit on weight - 1": {
			source:  NewNode("A"),
			maxW:    50,
			maxHops: 2,
			output: []*shortestPathTerm{
				&shortestPathTerm{NewNode("A"), 0},
				&shortestPathTerm{NewNode("B"), 20},
				&shortestPathTerm{NewNode("F"), 30},
				&shortestPathTerm{NewNode("G"), 30},
			},
		},
		"shortest path with 3 hops and no limit on weight": {
			source:  NewNode("A"),
			maxW:    300,
			maxHops: 3,
			output: []*shortestPathTerm{
				&shortestPathTerm{NewNode("A"), 0},
				&shortestPathTerm{NewNode("B"), 20},
				&shortestPathTerm{NewNode("C"), 40},
				&shortestPathTerm{NewNode("D"), 70},
				&shortestPathTerm{NewNode("F"), 30},
				&shortestPathTerm{NewNode("G"), 30},
			},
		},
		"shortest path with 4 hops and no limit on weight": {
			source:  NewNode("A"),
			maxW:    300,
			maxHops: 4,
			output: []*shortestPathTerm{
				&shortestPathTerm{NewNode("A"), 0},
				&shortestPathTerm{NewNode("B"), 20},
				&shortestPathTerm{NewNode("C"), 40},
				&shortestPathTerm{NewNode("D"), 50},
				&shortestPathTerm{NewNode("F"), 30},
				&shortestPathTerm{NewNode("G"), 30},
				&shortestPathTerm{NewNode("H"), 60},
			},
		},
		"shortest path with 4 hops and no limit on weight from a sink node": {
			source:  NewNode("H"),
			maxW:    300,
			maxHops: 4,
			output: []*shortestPathTerm{
				&shortestPathTerm{NewNode("H"), 0},
			},
		},
		"shortest path with 1 hops and no limit on weight from another node": {
			source:  NewNode("E"),
			maxW:    300,
			maxHops: 1,
			output: []*shortestPathTerm{
				&shortestPathTerm{NewNode("E"), 0},
				&shortestPathTerm{NewNode("G"), 30},
				&shortestPathTerm{NewNode("B"), 5},
			},
		},
		"shortest path with 2 hops and no limit on weight from another node": {
			source:  NewNode("E"),
			maxW:    300,
			maxHops: 2,
			output: []*shortestPathTerm{
				&shortestPathTerm{NewNode("E"), 0},
				&shortestPathTerm{NewNode("G"), 15},
				&shortestPathTerm{NewNode("F"), 15},
				&shortestPathTerm{NewNode("B"), 5},
			},
		},
		"shortest path with 2 hops and limit on weight from another node": {
			source:  NewNode("E"),
			maxW:    10,
			maxHops: 2,
			output: []*shortestPathTerm{
				&shortestPathTerm{NewNode("E"), 0},
				&shortestPathTerm{NewNode("B"), 5},
			},
		},
		"shortest path with 3 hops and no limit on weight from another node": {
			source:  NewNode("E"),
			maxW:    300,
			maxHops: 3,
			output: []*shortestPathTerm{
				&shortestPathTerm{NewNode("E"), 0},
				&shortestPathTerm{NewNode("G"), 15},
				&shortestPathTerm{NewNode("F"), 15},
				&shortestPathTerm{NewNode("C"), 25},
				&shortestPathTerm{NewNode("D"), 55},
				&shortestPathTerm{NewNode("B"), 5},
				&shortestPathTerm{NewNode("A"), 35},
			},
		},
	}
	for testName, test := range tests {
		suite.T().Logf("Running test case: %s", testName)
		sp, err := suite.g.Dijikstra(test.source, test.maxW, test.maxHops)
		assert.NoError(suite.T(), err)
		assert.ElementsMatch(suite.T(), sp, test.output)
	}

}

func TestShortestPathSuite(t *testing.T) {
	suite.Run(t, new(ShortestPathTestSuite))
}
