package utils

import (
	"context"
	"fmt"

	"github.com/AdarshGT33/ednr/adapters"
	"github.com/AdarshGT33/ednr/events"
	"github.com/redis/go-redis/v9"
)

// AttemptFallback tries the fallback channel before giving up to DLQ
func AttemptFallback(
	ctx context.Context,
	rdb *redis.Client,
	event events.Events,
	adapterMap map[string]adapters.NotificationAdapter,
	originalErr error,
) error {
	if event.FallbackChannel == "" {
		// No fallback defined, go straight to DLQ
		return MoveToDLQ(ctx, rdb, event, originalErr)
	}

	fallbackAdapter, exists := adapterMap[event.FallbackChannel]
	if !exists {
		fmt.Printf("⚠️ Fallback channel '%s' not found in adapters\n", event.FallbackChannel)
		return MoveToDLQ(ctx, rdb, event, originalErr)
	}

	fmt.Printf("⚡ [FALLBACK] Primary channel failed, attempting fallback via %s for user %s\n",
		event.FallbackChannel, event.User_ID)

	if err := fallbackAdapter.Send(event.Recipient, event.Message); err != nil {
		// Fallback also failed, now we go to DLQ
		fmt.Printf("❌ [FALLBACK] Fallback also failed: %v\n", err)
		return MoveToDLQ(ctx, rdb, event, err)
	}

	fmt.Printf("✅ [FALLBACK] Delivered via fallback channel %s for user %s\n",
		event.FallbackChannel, event.User_ID)
	return nil
}
