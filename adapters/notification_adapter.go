package adapters

type NotificationAdapter interface {
	Send(to, message string) error
}
