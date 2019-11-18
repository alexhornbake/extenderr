package extenderr

import (
	"fmt"
	"net/http"
	"github.com/pkg/errors"
	//"go.uber.org/zap"
)

const (
	SuperBadErrorCode = 1
)


func doSomethingElse() error {
	return errors.New("the root cause")
}

func doSomething() error {
	err := doSomethingElse()
	if err != nil {
		wrapped := errors.Wrap(WithErrorCode(err, SuperBadErrorCode), "something bad happened")
		return WithHttpStatus(wrapped, http.StatusInternalServerError)
	}
	return nil
}

func doSomethingWithTags() error {
	err := doSomething()
	if err != nil {
		return WithTags(err, "user", "beazley", "user_id", 666) 
	}
	return nil

}

func ExampleNamedCause() {
	err := doSomething()
	err = errors.Wrap(err, "more context")
	err2 := doSomethingWithTags()
	
	// Print the ints that we attached
	fmt.Println(ErrorCode(err), HttpStatus(err))

	// print the small version of the error (and chain)
	fmt.Println(err)

	// print the extended version with stack trace (skip in test)
	// fmt.Printf("%+v\n", err)

	// print the tags
	fmt.Printf("%+v", TagMap(err2))

	// Tags play nice with structured loggers (skip in test)
	/* logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	sugar.Infow(err2.Error(), Tags(err2)...)
	*/

	// Output: 1 500
	// more context: http status 500 : something bad happened: error code 1 : the root cause
	// map[user:beazley user_id:666]
}