/*******************************************************************************
 These wrapper functions exist as a convenience so that users of this package
 do not need to also import the official "errors" package should they need to
 use the wrapped functions.
*******************************************************************************/

package comerr

import "errors"

func New(text string) error {
	return errors.New(text)
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func Is(err error, target error) bool {
	return errors.Is(err, target)
}
