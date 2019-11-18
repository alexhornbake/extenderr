package main

import (
	"fmt"
	"errors"
	"net/http"

	"go.uber.org/zap"
	pkgErrors "github.com/pkg/errors"
	"github.com/alexhornbake/extenderr"
)

// Database/Model layer

// NoRows simulates an error from an external package like
// an SQL library.
type NoRows struct {}

func (e *NoRows) Error() string {
	return "no rows found"
}

// FindUser queries the database for a user that will never be found
func FindUser(userID uint64) (error) {
	err := &NoRows{}
	wrapped := pkgErrors.Wrap(err, "finding user")
	return extenderr.WithTags(wrapped, "user_id", userID)
}


// Controller/Handler layer

const (
	// UserNotFoundErrorCode is an error enum exposed to the client
	UserNotFoundErrorCode = 1
)

// HandleGetUser kinda/sorta handles a GET request for a user
func HandleGetUser() error {
	err := FindUser(2)
	if err != nil {
		// go 1.13 "Is" will check the entire error chain for an error
		if errors.Is(err, &NoRows{}) {
			wrapped := extenderr.WithErrorCode(err, UserNotFoundErrorCode)
			wrapped = extenderr.WithHttpStatus(err, http.StatusNotFound)
			wrapped = extenderr.WithHumanMessage(wrapped, "Sorry, we couldn't find that user")
			return wrapped
		}
		return err
	}
	return nil
}

// Router/Middleware layer

// HandleRequest shows how a router might call an opinionate handler
// how a logger might handle that error
// and how the opinionated handler might write the error response
func HandleRequest(logger *zap.SugaredLogger) {
	err := HandleGetUser()
	if err != nil {
		// Middleware can use the error message, tags, and stack trace to log a detailed message
		tags := extenderr.Tags(err)
		stackTrace := fmt.Sprintf("%+v", err)
		httpStatus := extenderr.HttpStatus(err)
		errorCode := extenderr.ErrorCode(err)
		humanMessage := extenderr.HumanMessage(err)

		tags = append(tags, "stack_trace", stackTrace, "http_status", httpStatus, "error_code", errorCode, "human_message", humanMessage)
		logger.Errorw(err.Error(), tags...)

		// the http response writer can put together a response that both a client and human can read
		// (and that is safe to expose to the end user)
		// care should be taken to avoid leaking internal error data to the client.
		fmt.Printf("\nHTTP Status: %d\nError Code: %d\nMessage: %s\n", httpStatus, errorCode, humanMessage)

		return
	}
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	HandleRequest(sugar)
	
}