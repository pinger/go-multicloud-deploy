package function

type Event struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func GetEvent(str string) Event {
	return Event{Message: "This is an event... " + str, Code: 9000}
}
