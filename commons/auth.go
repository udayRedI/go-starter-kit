package commons

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/udayRedI/go-starter-kit/lib"
)

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
