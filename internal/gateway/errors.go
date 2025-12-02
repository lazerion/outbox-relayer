package gateway

import (
	"errors"
	"fmt"
	"net/http"
)

// UpstreamError wraps an error from the SMS gateway with context about recoverability
type UpstreamError struct {
	Err         error
	StatusCode  int
	Recoverable bool
}

func (e *UpstreamError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("upstream error (status %d): %v", e.StatusCode, e.Err)
	}
	return fmt.Sprintf("upstream error: %v", e.Err)
}

// IsRecoverable checks whether an error can be retried
func IsRecoverable(err error) bool {
	if err == nil {
		return false
	}

	var ue *UpstreamError
	if errors.As(err, &ue) {
		return ue.Recoverable
	}

	return true
}

func isStatusRecoverable(code int) bool {
	return code == http.StatusTooManyRequests ||
		(code >= http.StatusInternalServerError &&
			code <= http.StatusNetworkAuthenticationRequired)
}

// WrapUpstreamError creates a standardized UpstreamError
func WrapUpstreamError(err error, statusCode int) *UpstreamError {
	return &UpstreamError{
		Err:         err,
		StatusCode:  statusCode,
		Recoverable: isStatusRecoverable(statusCode),
	}
}
