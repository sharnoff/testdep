package testdep

// Error is the wrapper type for specific errors that may be encountered. These are defined as
// global variables: NilFuncErr, CyclicDependencyErr, FunctionAlreadyPresentErr, and
// FunctionNotExecutedErr.
type Error struct{ string }

func (err Error) Error() string {
	return err.string
}

// These are the global errors that may be returned or panicked. FunctionAlreadyPresentErr and
// FunctionNotExecutedErr are both internal errors that are only included in case an unexpected
// condition is reached.
var (
	NilFuncErr                = Error{"Function is nil"}
	CyclicDependencyErr       = Error{"Graph has cyclic dependency"}
	FunctionAlreadyPresentErr = Error{"Internal error: function is already present in this graph"}
	FunctionNotExecutedErr    = Error{"Internal error: function was not executed"}
)
