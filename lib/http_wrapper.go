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
)

// Request stores information about HTTP request.
type Request struct {
	AppTitle         string
	Action           string
	Method           string
	RecordID         string
	FreeTextQuery    string
	scrollId         string
	Filter           map[string][]string
	Sort             []string
	Page             int
	PageSize         int
	Fields           []string
	Context          map[string]string
	UIReady          bool
	OtherQueryParams map[string][]string
	Body             io.ReadCloser
	Header           map[string][]string
	ID               string
	SentryContext    context.Context
	UserId           string
	isAuthenticated  bool
}

// DecodeBody unmarshals request JSON in a variable.
func (r *Request) DecodeBody(v interface{}) error {
	if r == nil || r.Body == nil {
		return errors.New("nothing to decode")
	}

	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(v)
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
	body, readErr := ioutil.ReadAll(r.Body)

	if readErr != nil {
		log.Println(fmt.Sprintf("%s: ERROR: Failed to ioutil.ReadAll: %s", r.ID, readErr))
		return nil, readErr
	}

	var data map[string]interface{}

	unmarshallErr := json.Unmarshal(body, &data)
	if unmarshallErr != nil {
		log.Println(fmt.Sprintf("%s: ERROR: Failed to unmarshal body: %s", r.ID, unmarshallErr))
		return nil, unmarshallErr
	}

	return &data, nil
}

// HasFilter checks whether a particular filter is set.
func (r *Request) HasFilter(filter string) bool {
	if len(r.Filter) == 0 {
		return false
	}

	_, ok := r.Filter[filter]
	return ok
}

// DefaultFilter sets default value for a filter is no value is specified.
func (r *Request) DefaultFilter(filter string, values ...string) {
	if len(r.Filter) == 0 {
		r.Filter = make(map[string][]string)
	}

	_, ok := r.Filter[filter]
	if ok {
		return
	}

	if values == nil {
		values = []string{""}
	}
	r.Filter[filter] = values
}

// GetFilter gets the specified filter's value. It returns nil when the filter doesn't exist.
func (r *Request) GetFilter(filter string) []string {
	if len(r.Filter) == 0 {
		return nil
	}

	return r.Filter[filter]
}

// SetFilter sets a filter value.
func (r *Request) SetFilter(filter string, values ...string) {
	if len(r.Filter) == 0 {
		r.Filter = make(map[string][]string)
	}

	if values == nil {
		values = []string{""}
	}

	r.Filter[filter] = values
}

// RemoveFilter removes a filter if it is present.
func (r *Request) RemoveFilter(filter string) {
	delete(r.Filter, filter)
}

// FilterEquals checks whether a given filter has a particular (single) value.
func (r *Request) FilterEquals(filter string, value string) bool {
	if len(r.Filter) == 0 {
		return false
	}

	values := r.Filter[filter]
	if len(values) != 1 {
		return false
	}

	return values[0] == value
}

// FilterContains checks whether a given filter has a particular value (among many possibly).
func (r *Request) FilterContains(filter string, value string) bool {
	if len(r.Filter) == 0 {
		return false
	}

	values := r.Filter[filter]
	for _, v := range values {
		if v == value {
			return true
		}
	}

	return false
}

// FilterAssertAndRemove ensures that if the filter is present it has a particular value and then removes it.
func (r *Request) FilterAssertAndRemove(filter string, value string) error {
	if len(r.Filter) == 0 {
		return nil
	}

	values, present := r.Filter[filter]
	if !present {
		return nil
	}

	if len(values) != 1 || values[0] != value {
		return errors.New("invalid value of filter: " + filter)
	}

	delete(r.Filter, filter)
	return nil
}

// FilterMultipleAssertAbsent determines whether provided filters are absent.
func (r *Request) FilterMultipleAssertAbsent(filters ...string) error {
	if len(r.Filter) == 0 {
		return nil
	}

	for _, filter := range filters {
		_, present := r.Filter[filter]
		if present {
			return errors.New("invalid filter: " + filter)
		}
	}

	return nil
}

// GetContext returns value of particular context type if it is provided in the request.
func (r *Request) GetContext(contextType string) string {
	if r == nil || r.Context == nil {
		return ""
	}

	return r.Context[contextType]
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

func (r *Request) GetFromQueryParam(key string) *string {
	if _, found := r.OtherQueryParams[key]; !found {
		return nil
	}

	if len(r.OtherQueryParams[key]) == 0 {
		return nil
	}

	return &r.OtherQueryParams[key][0]
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
