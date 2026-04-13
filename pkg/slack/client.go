package slack

import (
	"Aj-vrod/github-notifier/types"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Config struct {
	WebhookURL string `envconfig:"SLACK_WEBHOOK_URL" required:"true"`
}

type SlackClient struct {
	webhookURL string
	httpClient *http.Client
	cfg        *Config
}

func NewSlackClient(cfg *Config) *SlackClient {
	return &SlackClient{
		webhookURL: cfg.WebhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cfg: cfg,
	}
}

func (c *SlackClient) SendNotification(message string) error {
	msg := types.Message{
		Text: message,
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.webhookURL,
		"application/json",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack returned status %d", resp.StatusCode)
	}

	log.Println("Notification successfully sent to Slack")
	return nil
}
