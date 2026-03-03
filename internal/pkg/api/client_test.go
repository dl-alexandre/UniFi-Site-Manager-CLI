package api

import (
	"errors"
	"fmt"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		opts    ClientOptions
		wantErr bool
		errType string
	}{
		{
			name: "valid client",
			opts: ClientOptions{
				APIKey:  "test-api-key",
				BaseURL: "https://api.ui.com",
				Timeout: 30,
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			opts: ClientOptions{
				BaseURL: "https://api.ui.com",
				Timeout: 30,
			},
			wantErr: true,
			errType: "AuthError",
		},
		{
			name: "custom base URL",
			opts: ClientOptions{
				APIKey:  "test-api-key",
				BaseURL: "https://custom.ui.com",
				Timeout: 30,
			},
			wantErr: false,
		},
		{
			name: "default timeout",
			opts: ClientOptions{
				APIKey: "test-api-key",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if client == nil {
				t.Error("NewClient() returned nil client")
				return
			}
			if client.apiKey != tt.opts.APIKey {
				t.Errorf("client.apiKey = %v, want %v", client.apiKey, tt.opts.APIKey)
			}
		})
	}
}

func TestCalculateBackoff(t *testing.T) {
	client, _ := NewClient(ClientOptions{APIKey: "test"})

	tests := []struct {
		attempt int
		min     int64 // minimum expected in milliseconds
		max     int64 // maximum expected in milliseconds
	}{
		{attempt: 0, min: 0, max: 0},
		{attempt: 1, min: 1000, max: 2000},
		{attempt: 2, min: 2000, max: 3000},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("attempt_%d", tt.attempt), func(t *testing.T) {
			delay := client.calculateBackoff(tt.attempt)
			delayMs := delay.Milliseconds()
			if delayMs < tt.min || delayMs > tt.max {
				t.Errorf("calculateBackoff(%d) = %dms, want between %dms and %dms",
					tt.attempt, delayMs, tt.min, tt.max)
			}
		})
	}
}

func TestShouldRetry(t *testing.T) {
	client, _ := NewClient(ClientOptions{APIKey: "test"})

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"timeout error", errors.New("connection timeout"), true},
		{"connection refused", errors.New("connection refused"), true},
		{"no such host", errors.New("no such host"), true},
		{"temporary error", errors.New("temporary failure"), true},
		{"random error", errors.New("something else"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.shouldRetry(tt.err)
			if result != tt.expected {
				t.Errorf("shouldRetry() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
	}{
		{
			name:     "AuthError",
			err:      &AuthError{Message: "test"},
			wantCode: ExitAuthFailure,
		},
		{
			name:     "PermissionError",
			err:      &PermissionError{Message: "test"},
			wantCode: ExitPermissionDenied,
		},
		{
			name:     "NotFoundError",
			err:      &NotFoundError{Resource: "test"},
			wantCode: ExitValidationError,
		},
		{
			name:     "RateLimitError",
			err:      &RateLimitError{RetryAfter: 60},
			wantCode: ExitRateLimited,
		},
		{
			name:     "ValidationError",
			err:      &ValidationError{Message: "test"},
			wantCode: ExitValidationError,
		},
		{
			name:     "NetworkError",
			err:      &NetworkError{Message: "test"},
			wantCode: ExitNetworkError,
		},
		{
			name:     "nil error",
			err:      nil,
			wantCode: ExitSuccess,
		},
		{
			name:     "unknown error",
			err:      errors.New("unknown"),
			wantCode: ExitGeneralError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetExitCode(tt.err)
			if got != tt.wantCode {
				t.Errorf("GetExitCode() = %d, want %d", got, tt.wantCode)
			}
		})
	}
}
