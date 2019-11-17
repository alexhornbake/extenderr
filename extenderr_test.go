package extenderr

import (
	"fmt"
	"github.com/pkg/errors"
)

const superBadName = "SUPER_BAD"


func doSomethingElse() error {
	return errors.New("the root cause")
}

func doSomething() error {
	err := doSomethingElse()
	if err != nil {
		return errors.Wrap(NewNamed(err, superBadName), "something bad happened")
	}
	return nil
}

func doSomethingWithTags() error {
	err := doSomething()
	if err != nil {
		return NewTags(err, "user", "beazley", "user_id", 666) 
	}
	return nil

}

func ExampleNamedCause() {
	
	err := doSomething()
	err = errors.Wrap(err, "more context")
	err2 := doSomethingWithTags()
	
	fmt.Println(
		// check if any error in the cause chain is named superbad.
		IsAnyCauseNamed(err, superBadName),
	)

	fmt.Println(
		IsRootCauseNamed(err, superBadName),
	)

	fmt.Printf("%s\n", err)

	fmt.Printf("%+v", GetTagMap(err2))

	// Output: true
	// false
	// more context: something bad happened: SUPER_BAD: the root cause
	// map[user:beazley user_id:666]
}