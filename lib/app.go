package lib

import "strings"

type QueueAction func(string, string, string) (func(string), error)
type QueueRoute map[string]QueueAction

type HttpMethod string

const (
	GET    HttpMethod = "GET"
	HEAD   HttpMethod = "HEAD"
	POST   HttpMethod = "POST"
	PUT    HttpMethod = "PUT"
	PATCH  HttpMethod = "PATCH"
	DELETE HttpMethod = "DELETE"
)

type HttpAction struct {
	Action         string
	Method         HttpMethod
	Handler        func(*Request) *Response
	AuthValidators []AuthValidatorCallback
}

func (httpAction *HttpAction) Validate() {
	if !StringLenGtZero(string(httpAction.Method)) {
		httpAction.Method = GET
	}

	httpAction.Action = strings.TrimSuffix(httpAction.Action, "/")
	httpAction.Action = strings.TrimPrefix(httpAction.Action, "/")
}

// type MethodRoute map[HttpMethod]HttpAction

type HttpRoute map[string]map[HttpMethod]HttpAction

type App interface {
	Title() string
	Init(service *Service)
	Routes() []HttpAction
	QueueHandlers() QueueRoute
}
