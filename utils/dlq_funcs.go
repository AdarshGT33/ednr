package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AdarshGT33/ednr/events"
	"github.com/redis/go-redis/v9"
)

const (
	DLQ_QUEUE_KEY   = "dlq"
	RETRY_QUEUE_KEY = "retry_queue"
)

// move failed events to dlq
func MoveToDLQ(ctx context.Context, rdb *redis.Client, event events.Events, finalError error) error {
	event.LastError = finalError.Error()
	event.LastAttemptAt = time.Now()

	eventJson, err := json.Marshal(event)
	if err != nil {
		fmt.Printf("Failed to marshal event to dead letter queue: %w\n", err)
	}

	if err := rdb.LPush(ctx, DLQ_QUEUE_KEY, eventJson).Err(); err != nil {
		return fmt.Errorf("Failed to push to dlq: %w\n", err)
	}

	fmt.Printf("💀 [DLQ] Event moved to dead letter queue after %d attempts: %s\n",
		event.AttemptCount, event.Event_Type)

	// for production, could send alert to slack for dlq

	return nil
}

// Schedule event for retry with backoff
func ScheduleRetry(ctx context.Context, rdb *redis.Client, event events.Events, err error) error {
	event.AttemptCount++
	event.LastError = err.Error()
	event.LastAttemptAt = time.Now()

	backoff := event.GetBackOffDuration()

	eventJSON, marshalErr := json.Marshal(event)
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal event for retry: %w", marshalErr)
	}

	// Push to retry queue
	if pushErr := rdb.LPush(ctx, RETRY_QUEUE_KEY, eventJSON).Err(); pushErr != nil {
		return fmt.Errorf("failed to push to retry queue: %w", pushErr)
	}

	fmt.Printf("🔄 [RETRY] Scheduling retry #%d for event %s in %v\n",
		event.AttemptCount, event.Event_Type, backoff)

	return nil
}

// Get DLQ statistics
func GetDLQStats(ctx context.Context, rdb *redis.Client) (int64, error) {
	return rdb.LLen(ctx, DLQ_QUEUE_KEY).Result()
}

// Retrieve events from DLQ for manual inspection
func ListDLQEvents(ctx context.Context, rdb *redis.Client, limit int64) ([]events.Events, error) {
	results, err := rdb.LRange(ctx, DLQ_QUEUE_KEY, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	eventss := make([]events.Events, 0, len(results))
	for _, result := range results {
		var event events.Events
		if err := json.Unmarshal([]byte(result), &event); err == nil {
			eventss = append(eventss, event)
		}
	}

	return eventss, nil
}
