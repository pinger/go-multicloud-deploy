package function

type Event struct {
	Message string
	Code    int
}

func GetEvent() Event {
	return Event{Message: "This is an event...", Code: 9000}
}
