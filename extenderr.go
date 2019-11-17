package extenderr

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
)

// Name is an error name
type Name string

// extended is an error that has a name, and key/value pairs attached to it.
// Sentinel errors are frowned upon because they provide little extra detail.
// extended errors have the same convienence, but implement the Cause/Unwrap
// interface so that a named error can be inserted in to an error chain with
// extra context added.
type extended struct {
	error
	name Name
	keysAndValues []interface{}
}

type tagged interface {
	Tags() []interface{}
}

// Name returns the name of the error
func (e *extended) Name() Name { return e.name }

// Tags returns the slice of keys and values attached to this error
func (e *extended) Tags() []interface{} { return e.keysAndValues }

// Cause returnd the cause of the error
func (e *extended) Cause() error { return e.error }

// Unwrap provides compatibility with go1.13 errors
func (e * extended) Unwrap() error { return e.error}

// Error implements the Error interface
func (e *extended) Error() string {
	return string(e.name) + ": " + e.Cause().Error()
}

// Format implements the Formatter interface
func (e *extended) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", e.Cause())
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, string(e.name))
	}
}

// NewNamed creates a new error with a name that can be checked, and an optional
// set of key/value tags attached
func NewNamed(err error, name Name, keysAndValues ...interface{}) error {
	return &extended{
		error: err,
		name: name,
		keysAndValues: keysAndValues,
	}
}

// NewTags creates an unnamed error with a set of key/value tags attached.
func NewTags(err error, keysAndValues ...interface{}) error {
	return &extended{
		error: err,
		keysAndValues: keysAndValues,
	}
}

// IsNamed returns true if this specific error has a name
func IsNamed(err error, name Name) bool {
	if namedErr, ok := err.(*extended); ok {
		if namedErr.Name() == "" {
			return false
		}
		return namedErr.Name() == name
	}
	return false
}

// IsAnyCauseNamed checks if any errors in the error chain goes by this name
func IsAnyCauseNamed(errToWalk error, name Name) bool {
	return walkErrorChain(errToWalk, func(err error) bool {
		return IsNamed(err, name)
	})
}

// IsRootCauseNamed only returns true if the error at the root of the chain goes by this name
func IsRootCauseNamed(errToWalk error, name Name) bool {
	return IsNamed(errors.Cause(errToWalk), name)
}

// GetTags will return all tags attached to errors in the error chain.
// A slice of interface{} is a common paramter for structured loggers to accept.
func GetTags(errToWalk error) []interface{} {
	allTags := []interface{}{}
	walkErrorChain(errToWalk, func(err error) bool {
		if tagged, ok := err.(tagged); ok {
			allTags = append(allTags, tagged.Tags()...)
		}
		return false
	})
	return allTags
}

// GetTagMap will return a map of all tags in the error chain
// this is a best effort, unblanced key pairs will be made even,
// and duplicate tags overwritten (inner most tag wins)
func GetTagMap(errToWalk error) map[interface{}]interface{} {
	allTags := map[interface{}]interface{}{}
	walkErrorChain(errToWalk, func(err error) bool {
		if tagged, ok := err.(tagged); ok {
			tags := tagged.Tags()
			if len(tags) % 2 != 0{
				tags = append(tags, "unbalanced tag")
			}
			for i:=0; i<len(tags); i=i+2 {
				allTags[tags[i]] = tags[i+1]
			}
		}
		return false
	})
	return allTags
}

// errorIterator is a function that is intended to be used
// with a closure to walk the cause chain and collect/check things
// it should return true to stop walking the chain.
type errorIterator func(error) bool

// walkErrorChain will walk the error chain and run the
// errorIterator on every error in the chain unless
// the errorIterator returns true to signal an early return.
func walkErrorChain(err error, f errorIterator) bool {
	type causer interface {
		Unwrap() error
	}

	for err != nil {
		if f(err) {
			return true
		}
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Unwrap()
	}
	return f(err)
}
