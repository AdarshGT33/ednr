package events

import "time"

type Events struct {
	User_ID    string `json:"user_id"`
	Event_Type string `json:"event_type"`
	Message    string `json:"message"`
	Severity   string `json:"severity"`
	Recipient  string `json:"recipient"`

	AttemptCount  int       `json:"attempt_count"`
	MaxAttempts   int       `json:"max_attempts"`
	LastError     string    `json:"last_error"`
	CreatedAt     time.Time `json:"created_at"`
	LastAttemptAt time.Time `json:"last_attempt_at"`
}

func DetermineChannel(event Events) string {
	if event.Severity == "high" {
		return "sms"
	} else {
		return "email"
	}
}

func (e *Events) ShouldRetry() bool {
	return e.AttemptCount < e.MaxAttempts
}

func (e *Events) GetBackOffDuration() time.Duration {
	seconds := 1 << e.AttemptCount
	if seconds > 60 {
		seconds = 60
	}
	return time.Duration(seconds) * time.Second
}
