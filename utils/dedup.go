package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/AdarshGT33/ednr/events"
	"github.com/redis/go-redis/v9"
)

const DEDUP_TTL = 2 * time.Minute

func dedupKey(event events.Events) string {
	return fmt.Sprintf("dedup:%s:%s", event.User_ID, event.Event_Type)
}

// IsDuplicate checks if this event was already processed recently
func IsDuplicate(ctx context.Context, rdb *redis.Client, event events.Events) (bool, error) {
	key := dedupKey(event)
	exists, err := rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("dedup check failed: %w", err)
	}
	return exists > 0, nil
}

// MarkProcessed stamps the event in Redis so duplicates are caught within the TTL window
func MarkProcessed(ctx context.Context, rdb *redis.Client, event events.Events) error {
	key := dedupKey(event)
	if err := rdb.Set(ctx, key, "1", DEDUP_TTL).Err(); err != nil {
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}
	return nil
}
