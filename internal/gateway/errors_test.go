package gateway_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lazerion/outbox-relayer/internal/gateway"
)

func TestUpstreamError_Error(t *testing.T) {
	tests := []struct {
		name       string
		err        *gateway.UpstreamError
		wantString string
	}{
		{
			name: "with status code",
			err: &gateway.UpstreamError{
				Err:        errors.New("internal server error"),
				StatusCode: 500,
			},
			wantString: "upstream error (status 500): internal server error",
		},
		{
			name: "without status code",
			err: &gateway.UpstreamError{
				Err: errors.New("network failure"),
			},
			wantString: "upstream error: network failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.wantString {
				t.Errorf("Error() = %q, want %q", got, tt.wantString)
			}
		})
	}
}

func TestIsRecoverable(t *testing.T) {
	netErr := fmt.Errorf("network timeout")
	ueRecoverable := &gateway.UpstreamError{
		Err:         errors.New("server error"),
		StatusCode:  500,
		Recoverable: true,
	}
	ueNonRecoverable := &gateway.UpstreamError{
		Err:         errors.New("bad request"),
		StatusCode:  400,
		Recoverable: false,
	}

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"plain network error", netErr, true},
		{"recoverable UpstreamError", ueRecoverable, true},
		{"non-recoverable UpstreamError", ueNonRecoverable, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gateway.IsRecoverable(tt.err)
			if got != tt.want {
				t.Errorf("IsRecoverable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrapUpstreamError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		statusCode int
		want       bool
	}{
		{"500 status", errors.New("internal error"), 500, true},
		{"429 status", errors.New("too many requests"), 429, true},
		{"400 status", errors.New("bad request"), 400, false},
		{"200 status", errors.New("ok"), 200, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ue := gateway.WrapUpstreamError(tt.err, tt.statusCode)
			if ue.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %d, want %d", ue.StatusCode, tt.statusCode)
			}
			if ue.Recoverable != tt.want {
				t.Errorf("Recoverable = %v, want %v", ue.Recoverable, tt.want)
			}
			if !errors.Is(ue.Err, tt.err) {
				t.Errorf("Err = %v, want %v", ue.Err, tt.err)
			}
		})
	}
}
