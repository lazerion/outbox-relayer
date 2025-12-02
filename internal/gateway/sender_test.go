package gateway_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lazerion/outbox-relayer/internal/gateway"
	"github.com/lazerion/outbox-relayer/internal/model"
	"github.com/stretchr/testify/require"
)

func createMessage() model.Message {
	return model.Message{
		ID:          1,
		PhoneNumber: "+123456789",
		Content:     "Hello",
	}
}

func TestWebhookSender_Send(t *testing.T) {
	tests := []struct {
		name           string
		serverHandler  http.HandlerFunc
		authKey        string
		expectedErr    string
		expectedRespID string
		expectedStatus string
		timeout        time.Duration
		cancelContext  bool
	}{
		{
			name: "accepted",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusAccepted)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"messageId": "def456",
					"message":   "accepted",
				})
			},
			expectedRespID: "def456",
			expectedStatus: "accepted",
		},
		{
			name: "non accepted status",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`bad request`))
			},
			expectedErr: "unexpected status code: 400",
		},
		{
			name: "invalid JSON response",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write([]byte(`{invalid-json}`))
			},
			expectedErr: "decode response body",
		},
		{
			name: "context canceled",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(2 * time.Second)
				w.WriteHeader(http.StatusOK)
			},
			expectedErr:   "execute request",
			cancelContext: true,
		},
		{
			name: "timeout",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(200 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
			},
			expectedErr: "execute request",
			timeout:     50 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.serverHandler)
			defer ts.Close()

			timeout := 5 * time.Second
			if tt.timeout != 0 {
				timeout = tt.timeout
			}

			sender := gateway.NewWebhookSender(ts.URL, tt.authKey, timeout)
			msg := createMessage()

			ctx := context.Background()
			if tt.cancelContext {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			resp, err := sender.Send(ctx, msg)
			if tt.expectedErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErr)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, tt.expectedRespID, resp.MessageID)
				require.Equal(t, tt.expectedStatus, resp.Message)
			}
		})
	}
}
