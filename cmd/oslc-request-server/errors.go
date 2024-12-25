package main

import (
	"errors"
	"github.com/go-playground/validator/v10"
)

// translateValidationError is a helper function  to translate a validation error into a more user-friendly error.
// It is only intended to be used to return information to the operator of the binary, not the downstream users of the
// service provided by the binary.
//
// The function will translate validation errors into a message that is understandable to a user without development
// insights into the service. Additionally, the error returned will make it easy to cross-reference the error with
// the documentation.
//
// If the error is not a validation error, the error passes through a sort of wrapper to normalize potentially unknown
// errors.
func translateValidationError(err error) error {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		newErr := errors.New("invalid configuration")
		for _, e := range ve {
			newErr = errors.Join(newErr, errors.New(e.Error()))
		}
		return newErr
	}
	outer := errors.New("this is an unknown error related to the configuration and is almost certainly a bug. Please report this error")
	return errors.Join(outer, err)
}
