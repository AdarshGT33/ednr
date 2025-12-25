package events

type Events struct {
	User_ID    string `json:"user_id"`
	Event_Type string `json:"event_type"`
	Message    string `json:"message"`
}
