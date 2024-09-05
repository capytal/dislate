package errors

type EventErr interface {
	Error() string
	Event() string
	Reply() error
	Send() error
	Log()
	Join(...error) EventErr
}
