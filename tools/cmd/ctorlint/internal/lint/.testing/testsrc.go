package testsrc

import "regexp"

// valid: unexported
type valid struct {
}

// valid: unexported and nolint'd.
//nolint:ctors
type validNoLint struct {
}

// valid: exported w/ `T` ctor.
// some other comment!
type ValidT struct {
}

// valid: exported w/ `*T` ctor.
type ValidTPtr struct {
}

// valid: exported w/ `(T, error)` ctor.
type ValidTErr struct {
}

// valid: exported w/ `(*T, error)` ctor.
type ValidTPtrErr struct {
}

// valid: exported w/o ctor w/ nolint.
//nolint:ctors
type ValidNoLint struct {
}

// invalid: exported w/o ctor.
type InvalidNoCtor struct{}

// invalid: bad ctor.
type InvalidCtor struct{}

func NewStrPtrWithErr() (*string, error) { return nil, nil }

func NewValidT() ValidT { return ValidT{} }

func NewValidTErr() (ValidTErr, error) { return ValidTErr{}, nil }

func NewValidTPtr() *ValidTPtrErr { return nil }

func NewValidTPtrErr() (*ValidTPtr, error) { return nil, nil }

func NewRegexpPtr() *regexp.Regexp { return nil }

func NewInvalidCtor() error { return nil }
