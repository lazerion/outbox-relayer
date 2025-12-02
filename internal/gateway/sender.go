package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/lazerion/outbox-relayer/internal/model"
)

type SendResponse struct {
	MessageID string `json:"messageId"`
	Message   string `json:"message"`
}

// Sender defines the contract for sending messages to an external SMS Gateway.
type Sender interface {
	Send(ctx context.Context, message model.Message) (*SendResponse, error)
}

// WebhookSender implements the Sender interface for the example webhook.site.
type WebhookSender struct {
	Client  *http.Client
	URL     string
	AuthKey string
}

type webhookRequest struct {
	To      string `json:"to"`
	Content string `json:"content"`
}

func NewWebhookSender(url, authKey string, timeout time.Duration) Sender {
	return &WebhookSender{
		Client: &http.Client{
			Timeout: timeout,
		},
		URL:     url,
		AuthKey: authKey,
	}
}

func (s *WebhookSender) Send(ctx context.Context, message model.Message) (*SendResponse, error) {
	reqBody := webhookRequest{
		To:      message.PhoneNumber,
		Content: message.Content,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, WrapUpstreamError(fmt.Errorf("failed to marshal request body: %w", err), 0)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.URL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, WrapUpstreamError(fmt.Errorf("failed to create request: %w", err), 0)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if s.AuthKey != "" {
		req.Header.Set("api-key", s.AuthKey)
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, WrapUpstreamError(fmt.Errorf("failed to execute request: %w", err), 0)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return nil, WrapUpstreamError(
			fmt.Errorf("unexpected status code: %d", resp.StatusCode),
			resp.StatusCode,
		)
	}

	var response SendResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, WrapUpstreamError(fmt.Errorf("failed to decode response body: %w", err), resp.StatusCode)
	}

	return &response, nil
}
