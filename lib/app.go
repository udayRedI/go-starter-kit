package lib

type QueueAction func(string, string, string) (func(string), error)
type QueueRoute map[string]QueueAction

type HttpAction struct {
	Handler        func(*Request) *Response
	AuthValidators []AuthValidatorCallback
}

type HttpRoute map[string]HttpAction

type SseAction struct {
	Handler       func(*Request) (*chan (string), *chan (error), func(*Request))
	AuthValidator AuthValidatorCallback
}

type SseRoute map[string]SseAction

type App interface {
	Title() string
	Init(service *Service)
	Routes() HttpRoute
	QueueHandlers() QueueRoute
	SseHandlers() SseRoute
}
