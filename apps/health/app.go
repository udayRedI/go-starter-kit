package health

import "github.com/udayRedI/go-starter-kit/lib"

type Health struct {
	service *lib.Service
}

func New() *Health {
	return &Health{}
}

func (health *Health) Title() string {
	return "health"
}

func (health *Health) Init(service *lib.Service) {
	health.service = service
}

func (health *Health) Routes() lib.HttpRoute {
	return lib.HttpRoute{
		"get": lib.HttpAction{
			Handler: health.Get,
		},
	}
}

func (health *Health) QueueHandlers() lib.QueueRoute {
	return lib.QueueRoute{}
}

func (health *Health) SseHandlers() lib.SseRoute {
	return lib.SseRoute{}
}

func (health *Health) Get(req *lib.Request) *lib.Response {
	return lib.SuccessResponse("OK!")
}
