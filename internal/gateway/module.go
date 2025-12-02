package gateway

import (
	"go.uber.org/fx"

	"github.com/lazerion/outbox-relayer/internal/config"
)

func NewWebhookSenderProvider(cfg *config.Config) Sender {
	return NewWebhookSender(
		cfg.Webhook.Url,
		cfg.Webhook.AuthKey,
		cfg.Webhook.Timeout,
	)
}

var Module = fx.Module(
	"webhook.site",
	fx.Provide(NewWebhookSenderProvider),
)
