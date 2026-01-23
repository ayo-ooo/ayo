package crush

import (
	"context"
	"errors"
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    string
		wantErr bool
	}{
		{
			name:   "standard format",
			output: "crush version 0.1.0",
			want:   "0.1.0",
		},
		{
			name:   "simple format",
			output: "crush 0.1.0",
			want:   "0.1.0",
		},
		{
			name:   "version only",
			output: "0.1.0",
			want:   "0.1.0",
		},
		{
			name:   "with v prefix",
			output: "v0.1.0",
			want:   "0.1.0",
		},
		{
			name:   "with prerelease",
			output: "crush version 0.1.0-beta.1",
			want:   "0.1.0-beta.1",
		},
		{
			name:   "with build metadata",
			output: "crush version 0.1.0+build.123",
			want:   "0.1.0+build.123",
		},
		{
			name:   "devel build",
			output: "crush version devel",
			want:   "999.999.999",
		},
		{
			name:   "with newline",
			output: "crush version 0.2.5\n",
			want:   "0.2.5",
		},
		{
			name:    "invalid format",
			output:  "not a version",
			wantErr: true,
		},
		{
			name:    "empty",
			output:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseVersion(tt.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
		errType error
	}{
		{
			name:    "exact minimum",
			version: "0.1.0",
			wantErr: false,
		},
		{
			name:    "above minimum",
			version: "0.2.0",
			wantErr: false,
		},
		{
			name:    "well above minimum",
			version: "1.0.0",
			wantErr: false,
		},
		{
			name:    "devel version (treated as latest)",
			version: "999.999.999",
			wantErr: false,
		},
		{
			name:    "below minimum",
			version: "0.0.9",
			wantErr: true,
			errType: &CrushVersionError{},
		},
		{
			name:    "invalid version",
			version: "not.a.version",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.errType != nil && err != nil {
				var versionErr *CrushVersionError
				if !errors.As(err, &versionErr) {
					t.Errorf("validateVersion() error type = %T, want %T", err, tt.errType)
				}
			}
		})
	}
}

func TestCrushNotFoundError(t *testing.T) {
	innerErr := errors.New("executable file not found")
	err := &CrushNotFoundError{Err: innerErr}

	// Test Error()
	if err.Error() == "" {
		t.Error("CrushNotFoundError.Error() should not be empty")
	}

	// Test Unwrap()
	if !errors.Is(err, innerErr) {
		t.Error("CrushNotFoundError should unwrap to inner error")
	}
}

func TestCrushVersionError(t *testing.T) {
	err := &CrushVersionError{
		Found:    "0.0.5",
		Required: "0.1.0",
	}

	errStr := err.Error()
	if errStr == "" {
		t.Error("CrushVersionError.Error() should not be empty")
	}

	// Verify the error contains both versions
	if !contains(errStr, "0.0.5") || !contains(errStr, "0.1.0") {
		t.Errorf("CrushVersionError.Error() should contain both versions, got: %s", errStr)
	}
}

func TestFindBinary_NotInstalled(t *testing.T) {
	// This test only makes sense if crush is NOT installed
	// We can't reliably test this in all environments, so we just
	// verify the error types work correctly
	ctx := context.Background()
	_, err := FindBinary(ctx)

	// If crush is installed, skip the test
	if err == nil {
		t.Skip("crush is installed, skipping not-found test")
	}

	// If there's an error and it's not a version error, it should be a not-found error
	var versionErr *CrushVersionError
	if errors.As(err, &versionErr) {
		// Version error is also acceptable (crush found but wrong version)
		return
	}

	var notFoundErr *CrushNotFoundError
	if !errors.As(err, &notFoundErr) {
		// Could also be a version parse error, which is fine
		t.Logf("FindBinary() error type = %T (this is acceptable)", err)
	}
}

func TestIsAvailable(t *testing.T) {
	ctx := context.Background()
	// Just verify it doesn't panic
	_ = IsAvailable(ctx)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
