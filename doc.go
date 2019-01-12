// Package testdep is designed to allow simple establishing of dependent testing. This package
// should be used in cases where tests rely on other features to work and, as such, there would be
// no point in executing certain tests if those features were not passing.
//
// This package is mostly for small conveniences.
//
// Usage
//
// Because the focus of testdep is to create graphs of testing dependencies, the central type is
// the graph, created by
//
//	g := testdep.New()
//
// After creating the graph, dependencies can be added via the Require() method:
//
//	// Foo requires A, B, and C
//	g.Require(testFoo, testA, testB, testC)
//	g.Require(testA, testC, testD)
//
// In the case above, if any of testA, testB, or testC fail, testFoo will not be run. Tests are run
// with (*testing.T).Run(). The name supplied to Run can be set by Graph.Name
package testdep
