# go-starter-kit
A lightweight starter kit to begin Go projects without worrying about which libraries to choose. Go boasts an intriguing ecosystem, where there are no frameworks like in other languages. In Python, you have Django and FastAPI, while in Node.js, there's a new library and a new approach to accomplish the same task almost daily, if not a framework. Gophers prefer to keep it simple. Go is primarily used for building large-scale microservices

My aim with this project is to create a starter kit that prevents you from making the same mistakes I did. I'll handle the integration of the best libraries, leaving you free to focus on writing code with ease.


In the future, I will also delve deeper into the 'why' behind the underlying library, which will help you better appreciate my choices. If you think there's something better, feel free to suggest otherwise.


# Table of Contents
- [Overview](#overview)
- [Routing](#routing)

## Overview <a name="overview"></a>
If you've ever worked with Django, Nest.js, or perhaps Angular, you're likely familiar with a modular application structure. This approach aids in keeping the codebase organized, maintainable, and scalable.

Every app needs to have a unique title and is crucial to ensure that cicular dependencies wont occur.

Every app should implement `App` interface

Example app looks as follows:
```

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

func (health *Health) Routes() []lib.HttpAction {
	return []lib.HttpAction{
		{
			Handler: health.Get,
		},
	}
}

func (health *Health) QueueHandlers() lib.QueueRoute {
	return lib.QueueRoute{}
}

func (health *Health) Get(req *lib.Request) *lib.Response {
	return lib.SuccessResponse("OK!")
}

```

App title is crucial for [routing](#routing).


## Routing <a name="routing"></a>
Used [net/http](https://pkg.go.dev/net/http) package as its the most basic router. 

Lets say you've an app `user` all routes that start with `/user` will be routed to the following function and everything after will be considered an action. In the following example, `/user/get` is a valid route where `get` is an action.

```
func (health *Health) Routes() []lib.HttpAction {
	return []lib.HttpAction{
		{
			Handler: health.Get,
			Method: lib.GET,
			Action: "get", // `/get` `/get/` `get/` will be translated to `health/get`
		},
	}
}
```

Every action accepts 4 attribute:
1. Handler: This is your controller which accpets http `Request` and is expected to return `Response` object
2. Method: This represents the http method and is optional, its defaulted to `GET` request
3. Action: An action that represents URL and is defaulted to empty string. No need to start or end with `/`, will be ignored if found.
4. AuthValidator: Will be covered in detail





