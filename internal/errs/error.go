package errs

import "errors"

var (
	ErrInvalidParam   = errors.New("[kuryr] invalid param")
	ErrInvalidStatus  = errors.New("[kuryr] invalid status")
	ErrRecordNotFound = errors.New("[kuryr] record not found")

	ErrNoActivatedTplVersion = errors.New("[kuryr] no activated channel template version")
	ErrNotApprovedTplVersion = errors.New("[kuryr] not approved channel template version")

	ErrFailedToSendNotification = errors.New("[kuryr] failed to send notification")
)
