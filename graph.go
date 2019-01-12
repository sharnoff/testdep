package testdep

import (
	"reflect"
	"testing"
)

type node struct {
	fn func(*testing.T)

	name string

	mapKey uintptr

	requires []*node

	// whether or not testing has finished
	done bool

	failed bool
}

// Graph is the central dependency graph constructed by its methods to help with package-level
// testing.
type Graph struct {
	// the uintptr key is the value returned by getKey
	all map[uintptr](*node)

	// do from 0 to n
	topoSorted []*node

	validated bool
}

// New constructs a new dependency graph
func New() *Graph {
	g := Graph{
		all: make(map[uintptr](*node)),
	}

	return &g
}

// Require indicates that the given function 'before' requires the following functions. None of
// these functions must have already been added to the graph -- if they are not recognized, they
// will be added.
//
// Note: Graphs can contain multiple sub-graphs that are not connected, in addition to lone
// functions with no requirements.
//
// If any given function is nil, Require will panic with NilFuncErr
func (g *Graph) Require(before func(*testing.T), require ...func(*testing.T)) {
	g.validated = false

	// 'before' node
	b := g.getNode(before)

	// required nodes
	rns := make([]*node, len(require))
	for i := range require {
		rns[i] = g.getNode(require[i])
	}

	// add all functions in 'require' that are not already required by the node
Outer:
	for _, r := range rns {
		for i := range b.requires {
			if b.requires[i] == r {
				continue Outer
			}
		}

		b.requires = append(b.requires, r)
	}
}

// Name provides a name to use when running subtests. Functions (even if a name isn't given) will
// simply be called via t.Run(name, fn).
//
// If the function has not already be added, this will add it it to the graph.
//
// If fn is nil, Name will panic with NilFuncErr.
func (g *Graph) Name(fn func(*testing.T), name string) {
	g.getNode(fn).name = name
	return
}

// NameAll calls Graph.Name on all given function-name pairs. This is provided for simplicity of
// syntax. Like Name, NameAll will panic with NilFuncErr if any of the given functions are nil
func (g *Graph) NameAll(pairs []struct {
	Fn   func(*testing.T)
	Name string
}) {
	for _, pair := range pairs {
		g.Name(pair.Fn, pair.Name)
	}
}

func (g *Graph) getNode(fn func(*testing.T)) *node {
	key := getKey(fn)
	if n := g.all[key]; n != nil {
		return n
	}

	return g.blank(fn)
}

func (g *Graph) blank(fn func(*testing.T)) *node {
	key := getKey(fn)

	// assume fn is not already present. If not, panic
	if n := g.all[key]; n != nil {
		panic(FunctionAlreadyPresentErr)
	}

	n := &node{
		fn:     fn,
		mapKey: key,
	}

	g.all[key] = n
	return n
}

func getKey(fn func(*testing.T)) uintptr {
	if fn == nil {
		panic(NilFuncErr)
	}

	return reflect.ValueOf(fn).Pointer()
}

// this is just for use in testing this package. It will always remain false
var ignoreNilTesting bool = false

// Test is the method called to perform all of the tests that are present in the Graph. The typical
// usage of this package finishes by calling Graph.Test()
func (g *Graph) Test(t *testing.T) error {
	if !g.validated {
		if err := g.Validate(); err != nil {
			return err
		}
	}

	for _, n := range g.topoSorted {
		// check all sub-nodes are done. If any failed, don't do this
		for _, in := range n.requires {
			if in.failed {
				n.failed = true
			}

			if !in.done {
				panic(FunctionNotExecutedErr)
			}
		}

		if n.failed {
			t.Logf("Function %q (%v) had requirements fail", n.name, n.fn)

			n.done = true
			continue
		}

		// actually run the function. If we're testing this package, perform a slightly different
		// operation
		if !ignoreNilTesting || t != nil {
			n.failed = !t.Run(n.name, n.fn)
		} else {
			// if we're testing:
			n.fn(nil)
		}

		n.done = true
	}

	return nil
}

// Validate is a subroutine that is performed before Graph.Test is run. Validate will only return
// an error if there is a cycle in the graph (i.e. there are functions that indirectly require
// themselves). If there is a cycle, Validate will return CyclicDependencyErr.
func (g *Graph) Validate() error {
	// performs a topoligical sort of the nodes, and puts the result into g.topoSorted
	// note: this is actually an inverted topological sort, pretending that all of the edges of the
	// graph run from requiree to requirer, instead of the other way around
	//
	// taken from https://en.wikipedia.org/wiki/Topological_sorting#Kahn's_algorithm

	g.topoSorted = make([]*node, 0, len(g.all))

	// stillHas contains, for each node 'n', all nodes 'in' that it requires
	stillHas := make(map[*node]map[*node]bool)

	// noDep contains all of the nodes that have no dependencies
	noDep := make(map[*node]bool)

	for _, n := range g.all {
		if len(n.requires) == 0 {
			noDep[n] = true
		} else {
			stillHas[n] = make(map[*node]bool)
			for _, in := range n.requires {
				stillHas[n][in] = true
			}
		}
	}

	for len(noDep) != 0 {
		// pluck a node from noDep, add it to g.topoSorted
		var n *node
		{
			for n = range noDep {
				break
			}
			delete(noDep, n)
			g.topoSorted = append(g.topoSorted, n)
		}

		for _, m := range g.all {
			if stillHas[m][n] {
				delete(stillHas[m], n)

				if len(stillHas[m]) == 0 {
					delete(stillHas, m)
					noDep[m] = true
				}
			}
		}
	}

	if len(stillHas) != 0 {
		return CyclicDependencyErr
	}

	g.validated = true
	return nil
}
