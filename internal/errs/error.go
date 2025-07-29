package errs

import "errors"

var (
	ErrInvalidParam   = errors.New("[kuryr] invalid param")
	ErrInvalidStatus  = errors.New("[kuryr] invalid status")
	ErrRecordNotFound = errors.New("[kuryr] record not found")
)
