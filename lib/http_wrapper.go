package lib

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

// Request stores information about HTTP request.
type Request struct {
	ID            string
	AppTitle      string
	Path          string
	Method        string
	Body          io.ReadCloser
	Header        map[string][]string
	SentryContext context.Context
	UserId        string //Fix this
	Auth          Auth
	Query         url.Values
}

func (r *Request) GetDecodedBody(data interface{}) error {
	body, readErr := ioutil.ReadAll(r.Body)

	if readErr != nil {
		log.Println(fmt.Sprintf("%s: ERROR: Failed to ioutil.ReadAll: %s", r.ID, readErr))
		return readErr
	}

	unmarshallErr := json.Unmarshal(body, data)
	if unmarshallErr != nil {
		log.Println(fmt.Sprintf("%s: ERROR: Failed to unmarshal body: %s", r.ID, unmarshallErr))
		return unmarshallErr
	}

	return nil
}

func (r *Request) GetBodyMap() (*map[string]interface{}, error) {

	var data map[string]interface{}
	return &data, r.GetDecodedBody(data)
}

func (r *Request) GetHeaderVal(key string) *string {
	if _, found := r.Header[key]; !found {
		return nil
	}
	auth := r.Header[key]
	if len(auth) == 0 {
		return nil
	}
	return &auth[0]
}

// Response stores information about HTTP response.
type Response struct {
	Status int
	Body   interface{}
}

// SuccessResponse is a utility function for responding upon success (HTTP status code 200).
func SuccessResponse(body interface{}) *Response {
	return &Response{
		Status: http.StatusOK,
		Body:   body,
	}
}

// SuccessResponseWithMessage is a utility function for responding upon success (HTTP status code 200).
func SuccessResponseWithMessage(msg string) *Response {
	return &Response{
		Status: http.StatusOK,
		Body:   "{\"Msg\": \"" + msg + "\"}",
	}
}

// NotFoundResponse is a utility function for responding with 404 HTTP status.
func NotFoundResponse() *Response {
	return &Response{
		Status: http.StatusNotFound,
		Body:   "{\"Msg\": \"doesn't exist\"}",
	}
}

// NotFoundResponse is a utility function for responding with 404 HTTP status.
func NotFoundResponseWithMessage(message string) *Response {
	return &Response{
		Status: http.StatusNotFound,
		Body:   fmt.Sprintf("{\"Msg\": %s}", message),
	}
}

// InvalidActionResponse creates response when an invalid action is requested.
func InvalidActionResponse() *Response {
	return &Response{
		Status: http.StatusNotFound,
		Body:   "{\"Msg\": \"invalid action\"}",
	}
}

// ErrorResponse is a utility function for responding during error states (HTTP status code 500).
func ErrorResponse(err error) *Response {
	CaptureSentryException(err.Error())
	return &Response{
		Status: http.StatusInternalServerError,
		Body:   "{\"Msg\": \"something went wrong on our side\"}",
	}
}

// ClientErrorResponse is similar to ErrorResponse but status 400 (client error).
// It sends the error string back to client as Msg.
func ClientErrorResponse(e error) *Response {
	return &Response{
		Status: http.StatusBadRequest,
		Body:   "{\"Msg\": \"" + e.Error() + "\"}",
	}
}

func AuthFailedResponse() *Response {
	return &Response{
		Status: http.StatusUnauthorized,
		Body:   "{\"Msg\": \"Auth failed, please try again\"}",
	}
}

// IsSuccess confirms whether a response is a success response.
func IsSuccess(resp *Response) bool {
	if resp == nil {
		return false
	}

	if resp.Status == http.StatusOK {
		return true
	}

	return false
}

// IsNotFound confirms whether the response is a 404.
func IsNotFound(resp *Response) bool {
	if resp == nil {
		return false
	}

	if resp.Status == http.StatusNotFound {
		return true
	}

	return false
}

func GetResp(apiResp interface{}, statusCode int, errMsg *string) *Response {

	switch statusCode {
	case http.StatusBadRequest:
		return ClientErrorResponse(errors.New(*errMsg))
	case http.StatusInternalServerError:
		return ErrorResponse(errors.New(*errMsg))
	case http.StatusNotFound:
		return NotFoundResponse()
	}

	return SuccessResponse(apiResp)
}
