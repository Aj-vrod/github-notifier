package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig_Success(t *testing.T) {
	// Set required environment variables
	os.Setenv("GITHUB_TOKEN", "test-token-123")
	os.Setenv("SLACK_WEBHOOK_URL", "https://hooks.slack.com/test")
	os.Setenv("POLL_INTERVAL", "45s")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("SLACK_WEBHOOK_URL")
		os.Unsetenv("POLL_INTERVAL")
	}()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}

	if cfg.GithubCfg.Token != "test-token-123" {
		t.Errorf("GithubCfg.Token = %v, want %v", cfg.GithubCfg.Token, "test-token-123")
	}

	if cfg.SlackCfg.WebhookURL != "https://hooks.slack.com/test" {
		t.Errorf("SlackCfg.WebhookURL = %v, want %v", cfg.SlackCfg.WebhookURL, "https://hooks.slack.com/test")
	}

	if cfg.PollerCfg.PollInterval != 45*time.Second {
		t.Errorf("PollerCfg.PollInterval = %v, want %v", cfg.PollerCfg.PollInterval, 45*time.Second)
	}
}

func TestLoadConfig_MissingGitHubToken(t *testing.T) {
	// Set only Slack webhook, missing GitHub token
	os.Setenv("SLACK_WEBHOOK_URL", "https://hooks.slack.com/test")
	defer os.Unsetenv("SLACK_WEBHOOK_URL")

	// Ensure GITHUB_TOKEN is not set
	os.Unsetenv("GITHUB_TOKEN")

	_, err := LoadConfig()
	if err == nil {
		t.Error("LoadConfig() error = nil, want error for missing GITHUB_TOKEN")
	}
}

func TestLoadConfig_MissingSlackWebhook(t *testing.T) {
	// Set only GitHub token, missing Slack webhook
	os.Setenv("GITHUB_TOKEN", "test-token")
	defer os.Unsetenv("GITHUB_TOKEN")

	// Ensure SLACK_WEBHOOK_URL is not set
	os.Unsetenv("SLACK_WEBHOOK_URL")

	_, err := LoadConfig()
	if err == nil {
		t.Error("LoadConfig() error = nil, want error for missing SLACK_WEBHOOK_URL")
	}
}

func TestLoadConfig_MissingBothRequired(t *testing.T) {
	// Ensure both required vars are not set
	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("SLACK_WEBHOOK_URL")

	_, err := LoadConfig()
	if err == nil {
		t.Error("LoadConfig() error = nil, want error for missing required variables")
	}
}

func TestLoadConfig_DefaultPollInterval(t *testing.T) {
	// Set required vars but not POLL_INTERVAL
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("SLACK_WEBHOOK_URL", "https://hooks.slack.com/test")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("SLACK_WEBHOOK_URL")
	}()

	// Ensure POLL_INTERVAL is not set to test default
	os.Unsetenv("POLL_INTERVAL")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}

	// Default should be 30s as defined in poller package
	expectedDefault := 30 * time.Second
	if cfg.PollerCfg.PollInterval != expectedDefault {
		t.Errorf("PollerCfg.PollInterval = %v, want default %v", cfg.PollerCfg.PollInterval, expectedDefault)
	}
}

func TestLoadConfig_CustomPollInterval(t *testing.T) {
	tests := []struct {
		name     string
		interval string
		want     time.Duration
	}{
		{"1 minute", "1m", 1 * time.Minute},
		{"10 seconds", "10s", 10 * time.Second},
		{"2 minutes", "2m", 2 * time.Minute},
		{"90 seconds", "90s", 90 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("GITHUB_TOKEN", "test-token")
			os.Setenv("SLACK_WEBHOOK_URL", "https://hooks.slack.com/test")
			os.Setenv("POLL_INTERVAL", tt.interval)
			defer func() {
				os.Unsetenv("GITHUB_TOKEN")
				os.Unsetenv("SLACK_WEBHOOK_URL")
				os.Unsetenv("POLL_INTERVAL")
			}()

			cfg, err := LoadConfig()
			if err != nil {
				t.Fatalf("LoadConfig() error = %v, want nil", err)
			}

			if cfg.PollerCfg.PollInterval != tt.want {
				t.Errorf("PollerCfg.PollInterval = %v, want %v", cfg.PollerCfg.PollInterval, tt.want)
			}
		})
	}
}

func TestLoadConfig_InvalidPollInterval(t *testing.T) {
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("SLACK_WEBHOOK_URL", "https://hooks.slack.com/test")
	os.Setenv("POLL_INTERVAL", "invalid")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("SLACK_WEBHOOK_URL")
		os.Unsetenv("POLL_INTERVAL")
	}()

	_, err := LoadConfig()
	if err == nil {
		t.Error("LoadConfig() error = nil, want error for invalid POLL_INTERVAL format")
	}
}
