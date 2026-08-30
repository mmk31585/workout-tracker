package db

import (
	"context"
	"time"

	"github.com/mmk31585/workout-tracker/internal/config"
)

func QueryTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc) {
	cfg := config.GetConfig()
	timeout := 5 * time.Second 
	if cfg != nil {
		timeout = time.Duration(cfg.DB.QueryTimeout) * time.Second
	}
	return context.WithTimeout(ctx, timeout)
}
