package api

import (
	"strings"
	"testing"
)

// TestValidatePRURL_ValidURLs tests validation of valid GitHub PR URLs
func TestValidatePRURL_ValidURLs(t *testing.T) {
	validURLs := []struct {
		name string
		url  string
	}{
		{
			name: "standard URL",
			url:  "https://github.com/facebook/react/pull/12345",
		},
		{
			name: "single digit PR number",
			url:  "https://github.com/my-org/my-repo/pull/1",
		},
		{
			name: "repo with dots",
			url:  "https://github.com/user/repo.name/pull/999",
		},
		{
			name: "owner with hyphens",
			url:  "https://github.com/my-org/repo/pull/100",
		},
		{
			name: "repo with underscores",
			url:  "https://github.com/owner/my_repo/pull/50",
		},
		{
			name: "repo with mixed special chars",
			url:  "https://github.com/owner/repo-name.test_123/pull/1",
		},
		{
			name: "max length owner (39 chars)",
			url:  "https://github.com/" + strings.Repeat("a", 39) + "/repo/pull/1",
		},
		{
			name: "max length repo (100 chars)",
			url:  "https://github.com/owner/" + strings.Repeat("a", 100) + "/pull/1",
		},
	}

	for _, tt := range validURLs {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePRURL(tt.url)
			if err != nil {
				t.Errorf("expected valid URL, got error: %v", err)
			}
		})
	}
}

// TestParsePRURL_ValidURLs tests parsing of valid GitHub PR URLs
func TestParsePRURL_ValidURLs(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedOwner  string
		expectedRepo   string
		expectedNumber int
	}{
		{
			name:           "standard URL",
			url:            "https://github.com/facebook/react/pull/12345",
			expectedOwner:  "facebook",
			expectedRepo:   "react",
			expectedNumber: 12345,
		},
		{
			name:           "single digit PR",
			url:            "https://github.com/owner/repo/pull/1",
			expectedOwner:  "owner",
			expectedRepo:   "repo",
			expectedNumber: 1,
		},
		{
			name:           "large PR number",
			url:            "https://github.com/test/project/pull/999999",
			expectedOwner:  "test",
			expectedRepo:   "project",
			expectedNumber: 999999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := ParsePRURL(tt.url)
			if err != nil {
				t.Fatalf("expected valid URL, got error: %v", err)
			}

			if info.Owner != tt.expectedOwner {
				t.Errorf("expected owner %q, got %q", tt.expectedOwner, info.Owner)
			}
			if info.Repo != tt.expectedRepo {
				t.Errorf("expected repo %q, got %q", tt.expectedRepo, info.Repo)
			}
			if info.Number != tt.expectedNumber {
				t.Errorf("expected number %d, got %d", tt.expectedNumber, info.Number)
			}
		})
	}
}

// TestValidatePRURL_InvalidProtocol tests validation of URLs with invalid protocol
func TestValidatePRURL_InvalidProtocol(t *testing.T) {
	invalidURLs := []struct {
		name        string
		url         string
		expectedErr string
	}{
		{
			name:        "http protocol",
			url:         "http://github.com/owner/repo/pull/123",
			expectedErr: errProtocolNotHTTPS,
		},
		{
			name:        "ftp protocol",
			url:         "ftp://github.com/owner/repo/pull/123",
			expectedErr: errProtocolNotHTTPS,
		},
		{
			name:        "no protocol",
			url:         "github.com/owner/repo/pull/123",
			expectedErr: errProtocolNotHTTPS,
		},
	}

	for _, tt := range invalidURLs {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePRURL(tt.url)
			if err == nil {
				t.Errorf("expected error, got nil")
				return
			}
			if err.Error() != tt.expectedErr {
				t.Errorf("expected error %q, got %q", tt.expectedErr, err.Error())
			}
		})
	}
}

// TestValidatePRURL_InvalidDomain tests validation of URLs with invalid domain
func TestValidatePRURL_InvalidDomain(t *testing.T) {
	invalidURLs := []struct {
		name        string
		url         string
		expectedErr string
	}{
		{
			name:        "github enterprise",
			url:         "https://github.enterprise.com/owner/repo/pull/123",
			expectedErr: errDomainNotGitHub,
		},
		{
			name:        "gitlab domain",
			url:         "https://gitlab.com/owner/repo/pull/123",
			expectedErr: errDomainNotGitHub,
		},
		{
			name:        "bitbucket domain",
			url:         "https://bitbucket.org/owner/repo/pull/123",
			expectedErr: errDomainNotGitHub,
		},
	}

	for _, tt := range invalidURLs {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePRURL(tt.url)
			if err == nil {
				t.Errorf("expected error, got nil")
				return
			}
			if err.Error() != tt.expectedErr {
				t.Errorf("expected error %q, got %q", tt.expectedErr, err.Error())
			}
		})
	}
}

// TestValidatePRURL_InvalidPathFormat tests validation of URLs with invalid path format
func TestValidatePRURL_InvalidPathFormat(t *testing.T) {
	invalidURLs := []struct {
		name        string
		url         string
		expectedErr string
	}{
		{
			name:        "missing repo",
			url:         "https://github.com/owner/pull/123",
			expectedErr: errInvalidPathFormat,
		},
		{
			name:        "issues instead of pull",
			url:         "https://github.com/owner/repo/issues/123",
			expectedErr: errInvalidPathFormat,
		},
		{
			name:        "missing pull segment",
			url:         "https://github.com/owner/repo/123",
			expectedErr: errInvalidPathFormat,
		},
		{
			name:        "too many path segments",
			url:         "https://github.com/owner/repo/pull/123/extra",
			expectedErr: errInvalidPathFormat,
		},
		{
			name:        "missing owner",
			url:         "https://github.com/pull/123",
			expectedErr: errInvalidPathFormat,
		},
		{
			name:        "empty path",
			url:         "https://github.com/",
			expectedErr: errInvalidPathFormat,
		},
	}

	for _, tt := range invalidURLs {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePRURL(tt.url)
			if err == nil {
				t.Errorf("expected error, got nil")
				return
			}
			if err.Error() != tt.expectedErr {
				t.Errorf("expected error %q, got %q", tt.expectedErr, err.Error())
			}
		})
	}
}

// TestValidatePRURL_InvalidOwner tests validation of invalid owner names
func TestValidatePRURL_InvalidOwner(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectedErr string
	}{
		{
			name:        "owner too long (40 chars)",
			url:         "https://github.com/" + strings.Repeat("a", 40) + "/repo/pull/123",
			expectedErr: errOwnerTooLong,
		},
		{
			name:        "owner with special chars",
			url:         "https://github.com/owner@name/repo/pull/123",
			expectedErr: errOwnerInvalidFormat,
		},
		{
			name:        "owner with underscore",
			url:         "https://github.com/owner_name/repo/pull/123",
			expectedErr: errOwnerInvalidFormat,
		},
		{
			name:        "owner with dot",
			url:         "https://github.com/owner.name/repo/pull/123",
			expectedErr: errOwnerInvalidFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePRURL(tt.url)
			if err == nil {
				t.Errorf("expected error, got nil")
				return
			}
			if err.Error() != tt.expectedErr {
				t.Errorf("expected error %q, got %q", tt.expectedErr, err.Error())
			}
		})
	}
}

// TestValidatePRURL_InvalidRepo tests validation of invalid repo names
func TestValidatePRURL_InvalidRepo(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectedErr string
	}{
		{
			name:        "repo too long (101 chars)",
			url:         "https://github.com/owner/" + strings.Repeat("a", 101) + "/pull/123",
			expectedErr: errRepoTooLong,
		},
		{
			name:        "repo with special chars",
			url:         "https://github.com/owner/repo@name/pull/123",
			expectedErr: errRepoInvalidFormat,
		},
		{
			name:        "repo with spaces",
			url:         "https://github.com/owner/repo name/pull/123",
			expectedErr: errRepoInvalidFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePRURL(tt.url)
			if err == nil {
				t.Errorf("expected error, got nil")
				return
			}
			if err.Error() != tt.expectedErr {
				t.Errorf("expected error %q, got %q", tt.expectedErr, err.Error())
			}
		})
	}
}

// TestValidatePRURL_InvalidPRNumber tests validation of invalid PR numbers
func TestValidatePRURL_InvalidPRNumber(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectedErr string
	}{
		{
			name:        "non-numeric PR number",
			url:         "https://github.com/owner/repo/pull/abc",
			expectedErr: errPRNumberInvalidFormat,
		},
		{
			name:        "zero PR number",
			url:         "https://github.com/owner/repo/pull/0",
			expectedErr: errPRNumberInvalidFormat,
		},
		{
			name:        "negative PR number",
			url:         "https://github.com/owner/repo/pull/-123",
			expectedErr: errPRNumberInvalidFormat,
		},
		{
			name:        "empty PR number",
			url:         "https://github.com/owner/repo/pull/",
			expectedErr: errInvalidPathFormat,
		},
		{
			name:        "decimal PR number",
			url:         "https://github.com/owner/repo/pull/123.45",
			expectedErr: errPRNumberInvalidFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePRURL(tt.url)
			if err == nil {
				t.Errorf("expected error, got nil")
				return
			}
			if err.Error() != tt.expectedErr {
				t.Errorf("expected error %q, got %q", tt.expectedErr, err.Error())
			}
		})
	}
}

// TestValidatePRURL_MalformedURLs tests validation of malformed URLs
func TestValidatePRURL_MalformedURLs(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "empty URL",
			url:  "",
		},
		{
			name: "invalid URL format",
			url:  "not a url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePRURL(tt.url)
			if err == nil {
				t.Errorf("expected error for malformed URL, got nil")
			}
		})
	}
}

// TestValidatePRURL_URLsWithQueryAndFragment tests that query params and fragments are accepted
func TestValidatePRURL_URLsWithQueryAndFragment(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "URL with query params",
			url:  "https://github.com/owner/repo/pull/123?tab=files",
		},
		{
			name: "URL with fragment",
			url:  "https://github.com/owner/repo/pull/123#discussion",
		},
		{
			name: "URL with both query and fragment",
			url:  "https://github.com/owner/repo/pull/123?tab=files#discussion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePRURL(tt.url)
			if err != nil {
				t.Errorf("expected valid URL (query params and fragments should be ignored), got error: %v", err)
			}
		})
	}
}
