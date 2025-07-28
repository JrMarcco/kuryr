package errs

import "errors"

var (
	ErrInvalidParam   = errors.New("[kuryr] invalid param")
	ErrRecordNotFound = errors.New("[kuryr] record not found")
)
