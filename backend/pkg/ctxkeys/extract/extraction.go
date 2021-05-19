package extract

import (
	"context"

	"github.com/gcp-kit/gcpine-gae-example/backend/pkg/config"
	"github.com/gcp-kit/gcpine-gae-example/backend/pkg/ctxkeys"
)

// GetConfig - returns *config.Config
func GetConfig(ctx context.Context) (*config.Config, bool) {
	cfg, ok := ctx.Value(ctxkeys.ConfigKey{}).(*config.Config)
	return cfg, ok
}
