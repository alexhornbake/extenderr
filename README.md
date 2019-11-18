[![Documentation](https://godoc.org/github.com/alexhornbake/extenderr?status.svg)](http://godoc.org/github.com/alexhornbake/extenderr)

# extenderr
extend/wrap errors with: tags, error codes, http status, and human message. Plays nice with pkg/errors and go1.13 Unwrap.

```
// Example:

err := errors.New("user not found")
err = extenderr.WithHttpStatus(err, http.StatusNotFound)
err = extenderr.WithErrorCode(err, UserNotFoundErrorCode) // an enum
err = extenderr.WithTags(err, "user_id", 1234)
err = extenderr.WithHumanMessage(err, "Oh Darn, Looks like that user could not be found.")

// Later down the road, return the error code, human message, and http status to the client
message := extenderr.HumanMessage(err)
errorCode := extenderr.ErrorCode(err)
httpStatus := extenderr.HttpStatus(err))

// log the full error, and tags. Works well with structured loggers that accept ...interface{}
logger.Error(err.Error(), extenderr.Tags())

// If you are using pkg/errors, then the stack trace is still available:
stackTrace := fmt.Sprintf("%+v", err)

```
