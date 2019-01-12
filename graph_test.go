package testdep

import (
	"fmt"
	"reflect"
	"testing"
)

func ls(ints ...int) []int {
	return ints
}

func construct(indices [][]int, fns []func(*testing.T)) (g *Graph, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
				return
			}

			panic(r)
		}
	}()

	g = New()
	for _, tab := range indices {
		f := fns[tab[0]]
		g.Name(f, fmt.Sprint(tab[0]))

		fs := make([]func(*testing.T), len(tab)-1)
		for i := range fs {
			fs[i] = fns[tab[i+1]]
			g.Name(fs[i], fmt.Sprint(tab[i+1]))
		}

		g.Require(f, fs...)
	}

	err = g.Validate()
	return
}

func TestValidate(t *testing.T) {

	// length: 5
	fns := []func(*testing.T){
		func(t *testing.T) {},
		func(t *testing.T) {},
		func(t *testing.T) {},
		func(t *testing.T) {},
		func(t *testing.T) {},
	}

	table := []struct {
		err     error
		indices [][]int
	}{
		{nil, [][]int{
			ls(0, 1, 2),
			ls(1, 2, 3),
			ls(2, 3, 4),
		}},
		{nil, [][]int{
			ls(0, 1, 2),
			ls(1, 3, 4),
			ls(2, 3, 4),
		}},

		{CyclicDependencyErr, [][]int{
			ls(0, 2, 3),
			ls(1, 2, 3),
			ls(2, 0),
		}},
		{CyclicDependencyErr, [][]int{
			ls(0, 1, 2),
			ls(1, 2, 3),
			ls(2, 3, 4),
			ls(3, 0, 1),
		}},
	}

	for _, tab := range table {
		_, err := construct(tab.indices, fns)
		if (err != nil) != (tab.err != nil) {
			t.Errorf("Whether or not error was returned did not match. Expected: %q. Got: %q. Constructor indices: %v.",
				tab.err, err, tab.indices)
		} else if !reflect.DeepEqual(err, tab.err) {
			t.Errorf("Unexpected error type returned. Expected: %q. Got: %q. Constructor indices: %v.",
				tab.err, err, tab.indices)
		}
	}
}

func TestTest(t *testing.T) {
	ignoreNilTesting = true

	// setup:

	var indices [][]int

	fns := make([]func(*testing.T), 5)

	var done []bool // has length 5, declared later at each iteration

	// we need to declare these functions (somewhat) individually, because otherwise they'll be
	// recognized as the same function
	{
		// gives a *slightly* different function for each index of fns
		theActualFunction := func(i int) {
			fmt.Println("done[")
			// find the index in indices that fns[i] is at
			iTabIndex := -1
			for in := range indices {
				if indices[in][0] == i {
					iTabIndex = in
					break
				}
			}

			if iTabIndex != -1 {
				for r := 1; r < len(indices[iTabIndex]); r++ {
					rI := indices[iTabIndex][r]

					if !done[rI] {
						t.Errorf("fns[%d] required fns[%d], but [%d] had not been done. Constructor indices: %v",
							i, rI, rI, indices)
					}
				}
			}

			if done[i] {
				t.Errorf("fns[%d] was called more than once. Constructor indices: %v",
					i, indices)
			}

			done[i] = true
			return
		}

		fns[0] = func(dummy *testing.T) {
			theActualFunction(0)
		}

		fns[1] = func(dummy *testing.T) {
			theActualFunction(1)
		}

		fns[2] = func(dummy *testing.T) {
			theActualFunction(2)
		}

		fns[3] = func(dummy *testing.T) {
			theActualFunction(3)
		}

		fns[4] = func(dummy *testing.T) {
			theActualFunction(4)
		}
	}

	// establish result tables
	table := [][][]int{
		[][]int{
			ls(0, 1, 2, 3),
			ls(1, 2, 4),
			ls(3, 4),
		},
		[][]int{
			ls(0, 2, 3, 4),
			ls(1, 2, 3, 4),
			ls(2, 4),
			ls(3, 4),
		},
		[][]int{
			ls(0, 3, 4),
			ls(1, 2, 4),
			ls(2, 3, 4),
		},
	}

	for _, indices = range table {
		done = make([]bool, 5)
		shouldHaveFinished := make([]bool, len(fns))

		for _, l := range indices {
			for _, f := range l {
				shouldHaveFinished[f] = true
			}
		}

		g, err := construct(indices, fns)
		if err != nil {
			t.Errorf("Encountered error %q while constructing proper graph. Constructor indices: %v",
				err, indices)
		}

		// pass it nil because we don't actually care if the testing.T can do anything
		if err := g.Test(nil); err != nil {
			t.Errorf("Encountered error %q while testing proper graph. Constructor indices: %v Graph: %v",
				err, indices, g)
		} else {
			// check everything worked
			for i, shf := range shouldHaveFinished {
				if shf && !done[i] {
					t.Errorf("fns[%d] was not called but should have been. Constructor indices: %v", i, indices)
				}
			}
		}
	}
}
