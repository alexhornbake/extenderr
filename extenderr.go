// package extenderr is an error utility aimed at application servers.
// It's purpose is to supply the outter most caller (http handler, middleware, etc)
// eith useful info about the error, and communicate accurate and helpful status to clients and human end users.
//
// Gives errors additional annotations and retrivial of:
//     - a description that is safe to expose to humans (humanMessage interface)
//     - an enum error code that can be checked (errorCoder interface)
//     - an HTTP status code that can be returned (httpStatuser interface)
//     - key/value pairs that have been attached (tagger interface)
//
// All of the interfaces are private, but considered stable, such that if your
// use case deviates from this package, one should be able to implement the interface
// in a similar way that this package implements "Error", "Format", Cause", and "Unwrap".
//
// This package is safe to use on any error (and nil), it will return "zero" values for any unused
// fields, or any unimplimented interfaces when retrieving annotations.
package extenderr

import (
	"fmt"
	"io"
)

// wrapper is the go.13 errors interface for Unwrapping an error
// This interface is used to walk the error chain when necessary.
type wrapper interface {
	Unwrap() error
}

/*

 HumanMessage

*/

type humanMessage interface {
	HumanMessage() string
}

type withHumanMessage struct{
	error
	message string
}

// CustomerMessage returns a description of the error that is intended to be
// exposed to end users.
func (e *withHumanMessage) HumanMessage() string { return e.message }

// Unwrap returns the error that is being wrapped
func (e *withHumanMessage) Unwrap() error { return e.error}

// Error implements the Error interface
func (e *withHumanMessage) Error() string {
	return e.message + ": " + e.Unwrap().Error()
}

// Format implements the Formatter interface
func (e *withHumanMessage) Format(s fmt.State, verb rune) { format(e, s, verb) }

// WithHumanMessage wraps the error with a message intended for end users.
func WithHumanMessage(err error, message string) error {
	if err == nil {
		return err
	}
	return &withHumanMessage{
		error: err,
		message: message,
	}
}


// HumanMessage returns the first (outter most) message encountered in the error chain.
// The message is intended to be exposed to human. If not message exists or the error is nil
// it returns empty string.
func HumanMessage(errToWalk error) string {
	message := ""
	if errToWalk == nil {
		return message
	}
	walkErrorChain(errToWalk, func(err error) bool {
		if m, ok := err.(humanMessage); ok {
			message = m.HumanMessage()
		}
		return message != ""
	})
	return message
}

/*

 ErrorCode

*/

type errorCoder interface {
	ErrorCode() int
}

type withErrorCode struct{
	error
	errorCode int
}

// ErrorCode returns an error code enum that can be checked to identify an error
func (e *withErrorCode) ErrorCode() int { return e.errorCode }

// Unwrap returns the error that is being wrapped
func (e *withErrorCode) Unwrap() error { return e.error}

// Error implements the Error interface
func (e *withErrorCode) Error() string {
	return fmt.Sprintf("error code %d : ", e.errorCode) + e.Unwrap().Error()
}

// Format implements the Formatter interface
func (e *withErrorCode) Format(s fmt.State, verb rune) { format(e, s, verb) }

// WithErrorCode wraps the error with an error code for machines.
func WithErrorCode(err error, errorCode int) error {
	if err == nil {
		return err
	}
	return &withErrorCode{
		error: err,
		errorCode: errorCode,
	}
}

// ErrorCode returns the first (outter most) error code encountered in the error chain.
// An int enum error code is intended for signaling a specific error state to clients
// of an API.
func ErrorCode(errToWalk error) int {
	errorCode := 0
	if errToWalk == nil {
		return errorCode
	}
	walkErrorChain(errToWalk, func(err error) bool {
		if ec, ok := err.(errorCoder); ok {
			errorCode = ec.ErrorCode()
		}
		return errorCode != 0
	})
	return errorCode
}

/*

Http Status

*/

type httpStatuser interface {
	HttpStatus() int
}

type withHttpStatus struct{
	error
	httpStatus int
}

// HttpStatus returns an http status code that can be returned to a client
func (e *withHttpStatus) HttpStatus() int { return e.httpStatus }

// Unwrap returns the error that is being wrapped
func (e *withHttpStatus) Unwrap() error { return e.error}

// Error implements the Error interface
func (e *withHttpStatus) Error() string {
	return fmt.Sprintf("http status %d : ", e.httpStatus) + e.Unwrap().Error()
}

// Format implements the Formatter interface
func (e *withHttpStatus) Format(s fmt.State, verb rune) { format(e, s, verb) }

// WithHttpStatus wraps the error with an http status code.
func WithHttpStatus(err error, status int) error {
	if err == nil {
		return err
	}
	return &withHttpStatus{
		error: err,
		httpStatus: status,
	}
}

// HttpStatus returns the first (outter most) http status code encountered in the error chain.
func HttpStatus(errToWalk error) int {
	status := 0
	if errToWalk == nil {
		return status
	}
	walkErrorChain(errToWalk, func(err error) bool {
		if ws, ok := err.(httpStatuser); ok {
			status = ws.HttpStatus()
		}
		return status != 0
	})
	return status
}

/*

Tags

*/

type tagger interface {
	Tags() []interface{}
}

type withTags struct{
	error
	keysAndValues []interface{}
}

// Tags returns the slice of keys and values attached to this error
func (e *withTags) Tags() []interface{} { return e.keysAndValues }

// Unwrap returns the error that is being wrapped
func (e *withTags) Unwrap() error { return e.error}

// Error implements the Error interface
func (e *withTags) Error() string { return e.Unwrap().Error() }

// Format implements the Formatter interface
func (e *withTags) Format(s fmt.State, verb rune) { format(e, s, verb) }


// WithTags wraps an error with a set of key/value tags attached.
func WithTags(err error, keysAndValues ...interface{}) error {
	if err == nil {
		return err
	}
	return &withTags{
		error: err,
		keysAndValues: keysAndValues,
	}
}

// Tags returns all of the tags encountered in the error chain.
func Tags(errToWalk error) []interface{} {
	allTags := []interface{}{}
	walkErrorChain(errToWalk, func(err error) bool {
		if tagged, ok := err.(tagger); ok {
			allTags = append(allTags, tagged.Tags()...)
		}
		return false
	})
	return allTags
}

// TagMap will return a map of all tags in the error chain.
// This is a best effort, unblanced key pairs will be made even,
// and duplicate tags overwritten (the inner most tag wins).
func TagMap(errToWalk error) map[interface{}]interface{} {
	allTags := map[interface{}]interface{}{}
	walkErrorChain(errToWalk, func(err error) bool {
		if tagged, ok := err.(tagger); ok {
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

// helper to format a wrapped error
// compatible with pkg/errors "%+v" convention for stack traces
func format(err error, s fmt.State, verb rune) {
	w, ok := err.(wrapper)
	if !ok {
		io.WriteString(s, err.Error())
		return
	}
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", w.Unwrap())
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, err.Error())
	}
}

// errorIterator is a function that is intended to be used
// with a closure to walk the cause chain and collect/check things
// it should return true to stop walking the chain.
type errorIterator func(error) bool

// walkErrorChain will walk the error chain and run the
// errorIterator on every error in the chain unless
// the errorIterator returns true to signal an early return.
func walkErrorChain(err error, f errorIterator) bool {
	if err == nil {
		return false
	}
	for err != nil {
		if f(err) {
			return true
		}
		w, ok := err.(wrapper)
		if !ok {
			break
		}
		err = w.Unwrap()
	}
	return f(err)
}
