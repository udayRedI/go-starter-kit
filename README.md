# go-starter-kit
A lightweight starter kit to begin Go projects without worrying about which libraries to choose. Go boasts an intriguing ecosystem, where there are no frameworks like in other languages. In Python, you have Django and FastAPI, while in Node.js, there's a new library and a new approach to accomplish the same task almost daily, if not a framework. Gophers prefer to keep it simple. Go is primarily used for building large-scale microservices

My aim with this project is to create a starter kit that prevents you from making the same mistakes I did. I'll handle the integration of the best libraries, leaving you free to focus on writing code with ease.


In the future, I will also delve deeper into the 'why' behind the underlying library, which will help you better appreciate my choices. If you think there's something better, feel free to suggest otherwise.


Note: Do not make changes in lib folder unless you know what you're doing. 

# Table of Contents
- [Overview](#overview)
- [Routing](#routing)
- [Auth](#auth)
- [Cache](#cache)

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

Lets say you've an app `user`, all routes defined in `Routes()` are considered to be actions(sub-routes) of the same. In the following example, `/user/get` is a valid route where `get` is an action.

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


## Auth <a name="auth"></a>
Objective is not to provide auth but to help inject in routes. You will be able to reuse auth for every route.
Once auth is finalized, an `auth` object injected into `Request` and all controllers and services will have access to it.

Defining an auth:
```

type AuthPayload struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type RouteHeaderValidator struct {
	service *lib.Service
}

func NewRouteHeaderValidator() lib.AuthValidatorCallback {
	return func(service *lib.Service) lib.AuthValidator {
		return &RouteHeaderValidator{
			service: service,
		}
	}
}

func (rv *RouteHeaderValidator) Validate(req *lib.Request) lib.Auth {

	token := req.GetHeaderVal("Token")
	if token == nil {
		return lib.Auth{}
	}

	url := "https://api.example.com/user/"

	client := &http.Client{}
	extReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return lib.Auth{}
	}
	extReq.Header.Set("Authorization", *token)

	resp, err := client.Do(extReq)
	if err != nil {
		fmt.Println("Error making request:", err)
		return lib.Auth{}
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return lib.Auth{}
	}

	var user AuthPayload
	err = json.Unmarshal(body, &user)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return lib.Auth{}
	}

	return lib.Auth{
		IsAuthenticated: false,
		Payload:         user,
	}
}

```

Once you've defined an auth you want to stick it onto a route as so:


```
func (health *Health) Routes() []lib.HttpAction {
	return []lib.HttpAction{
		{
			Handler: health.Get,
			Method:  lib.GET,
			Action:  "get",
			AuthValidators: []lib.AuthValidatorCallback{
				commons.NewRouteHeaderValidator(),
			},
		},
	}
}
```

`AuthValidators` accepts multiple callbacks, which means you can attach multiple auth for a given route.



## Cache <a name="cache"></a>
Currently I've implemented only redis. So, if you're working with redis you're in luck. Chose [Go Redis](https://redis.uptrace.dev/) and its feature rich. Just ensure redis-redentials are passed json file in config folder.
```
"RedisCreds": {
	"Addr": "...",
	"Password": "...",
	"Db": 1
}
```
In order to access redis client inject using service `service.RedisClient`. [Go Redis](https://redis.uptrace.dev/) implements pooling so any operation you do would automatically close connection, one exception to this is redis.PubSub or redis.Conn, [link](https://redis.uptrace.dev/guide/go-redis-debugging.html#connection-pool-size).

In the future plan is to support multiple caches like memcached and more.