package intermediaries_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Kaese72/finding-registry/internal/intermediaries"
	"github.com/Kaese72/finding-registry/rest/apierrors"
)

func TestReportLocatorValidateError(t *testing.T) {
	// Test cases for ReportLocator.Validate
	var tests = []struct {
		locator intermediaries.ReportLocator
		err     apierrors.APIError
	}{
		// Missing values
		{
			intermediaries.ReportLocator{Type: intermediaries.IPv4, Value: "192.168.0.1", Distinguisher: ""},
			apierrors.APIError{Code: 400, WrappedError: fmt.Errorf("missing Distinguisher")},
		},
		{
			intermediaries.ReportLocator{Type: intermediaries.IPv4, Value: "", Distinguisher: "global"},
			apierrors.APIError{Code: 400, WrappedError: fmt.Errorf("missing Value")},
		},
		{
			intermediaries.ReportLocator{Type: "", Value: "192.168.0.1", Distinguisher: "global"},
			apierrors.APIError{Code: 400, WrappedError: fmt.Errorf("missing Type")},
		},
		// IPv4 validation
		{
			intermediaries.ReportLocator{Type: intermediaries.IPv4, Value: "192.168.0.1", Distinguisher: "global"},
			apierrors.APIError{Code: 400, WrappedError: fmt.Errorf("private IPv4 address cannot have a global distinguisher")},
		},
		{
			intermediaries.ReportLocator{Type: intermediaries.IPv4, Value: "192.168.0", Distinguisher: "global"},
			apierrors.APIError{Code: 400, WrappedError: fmt.Errorf("invalid IPv4 address: 192.168.0")},
		},
		{
			intermediaries.ReportLocator{Type: intermediaries.IPv4, Value: "127.0.0.1", Distinguisher: "global"},
			apierrors.APIError{Code: 400, WrappedError: fmt.Errorf("loopback IPv4 address not allowed")},
		},
		{
			intermediaries.ReportLocator{Type: intermediaries.IPv4, Value: "127.0.0.1", Distinguisher: "somethingspecial"},
			apierrors.APIError{Code: 400, WrappedError: fmt.Errorf("loopback IPv4 address not allowed")},
		},
		{
			intermediaries.ReportLocator{Type: intermediaries.Hostname, Value: "localhost", Distinguisher: "global"},
			apierrors.APIError{Code: 400, WrappedError: fmt.Errorf("hostname may not be 'localhost'")},
		},
	}
	for _, testInput := range tests {
		t.Run(testInput.locator.Value, func(t *testing.T) {
			err := testInput.locator.Validate()
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != testInput.err.Error() {
				t.Fatalf("expected error %v, got %v", testInput.err, err)
			}
			var apiErr apierrors.APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("expected APIError, got %T", err)
			}
			if apiErr.Code != testInput.err.Code {
				t.Fatalf("expected code %d, got %d", testInput.err.Code, apiErr.Code)
			}
			if apiErr.WrappedError.Error() != testInput.err.WrappedError.Error() {
				t.Fatalf("expected wrapped error %v, got %v", testInput.err.WrappedError, apiErr.WrappedError)
			}
		})
	}
}

func TestReportLocatorValidatePositive(t *testing.T) {
	// Test cases for ReportLocator.Validate
	var tests = []intermediaries.ReportLocator{
		{Type: intermediaries.IPv4, Value: "84.84.84.84", Distinguisher: "global"},
		{Type: intermediaries.HTTP, Value: "https://example.com", Distinguisher: "global"},
	}
	for _, testInput := range tests {
		t.Run(testInput.Value, func(t *testing.T) {
			err := testInput.Validate()
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
