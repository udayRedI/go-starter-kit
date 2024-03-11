package lib

import (
	"net/http"
	"net/url"
	"testing"
)

type MockResponseWriter struct {
	dataWritten string
	mockHeader  http.Header
	statusCode  int16
}

func (m *MockResponseWriter) Header() http.Header {
	if m.mockHeader == nil {
		m.mockHeader = make(http.Header)
	}
	return m.mockHeader
}

func (m *MockResponseWriter) Write(writtenBytes []byte) (int, error) {
	m.dataWritten = string(writtenBytes)
	return 0, nil
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = int16(statusCode)
}

func (m *MockResponseWriter) GotExpResp(exp string) bool {
	return m.dataWritten == exp
}

type MockApp struct {
}

func (mApp *MockApp) Title() string {
	return "mock-app"
}

func (mApp *MockApp) Init(s *Service) {
}

var AUTH_UID string = "UID1"
var AUTH_UID2 string = "UID2"

func (mApp *MockApp) Routes() []HttpAction {
	return []HttpAction{
		{
			Method: "GET",
			Action: "public-success-get",
			Handler: func(*Request) *Response {
				return SuccessResponse("Success")
			},
		},
		{
			Method:         "GET",
			Action:         "private-success-get",
			AuthValidators: []AuthValidatorCallback{NewMockAuthValidator(true, &AUTH_UID)},
			Handler: func(*Request) *Response {
				return SuccessResponse("Success")
			},
		},
		{
			Method:         "GET",
			Action:         "private-success-uid2",
			AuthValidators: []AuthValidatorCallback{NewMockAuthValidator(true, &AUTH_UID2)},
			Handler: func(*Request) *Response {
				return SuccessResponse("Success")
			},
		},
		{
			Method:         "GET",
			Action:         "private-success-with-no-uid",
			AuthValidators: []AuthValidatorCallback{NewMockAuthValidator(true, nil)},
			Handler: func(*Request) *Response {
				return SuccessResponse("Success")
			},
		},
		{
			Method:         "GET",
			Action:         "private-failure-get",
			AuthValidators: []AuthValidatorCallback{NewMockAuthValidator(false, nil)},
			Handler: func(*Request) *Response {
				return SuccessResponse("Success")
			},
		},
		{
			Method: "GET",
			Action: "models",
			Handler: func(*Request) *Response {
				return SuccessResponse("Success GET")
			},
		},
		{
			Method: "POST",
			Action: "models",
			Handler: func(*Request) *Response {
				return SuccessResponse("Success POST")
			},
		},
		{
			Action: "get",
			Handler: func(*Request) *Response {
				return SuccessResponse("Success GET")
			},
		},
		{
			Handler: func(*Request) *Response {
				return SuccessResponse("Success GET")
			},
		},
		{
			Method: "POST",
			Handler: func(*Request) *Response {
				return SuccessResponse("Success POST")
			},
		},
	}
}

func (mApp *MockApp) QueueHandlers() QueueRoute {
	return QueueRoute{}
}

func TestRouter(t *testing.T) {

	type input struct {
		title         string
		req           http.Request
		resp          http.ResponseWriter
		expResp       string
		expStatusCode int16
		apps          []App
	}

	inputs := []input{
		{
			title: "Options returns empty response",
			req: http.Request{
				Method: "OPTIONS",
			},
			resp:          &MockResponseWriter{},
			expStatusCode: 0,
			expResp:       "",
		},
		{
			title: "Public valid get API should return 200",
			req: http.Request{
				Method: "GET",
				URL: &url.URL{
					Path: "mock-app/public-success-get",
				},
			},
			expStatusCode: 200,
			resp:          &MockResponseWriter{},
			expResp:       "Success",
			apps:          []App{&MockApp{}},
		},
		{
			title: "Public invalid route should return 404",
			req: http.Request{
				Method: "GET",
				URL: &url.URL{
					Path: "mock-app/public-invalid-get",
				},
			},
			expStatusCode: 404,
			resp:          &MockResponseWriter{},
			expResp:       "{\"Msg\": \"doesn't exist\"}",
			apps:          []App{&MockApp{}},
		},
		{
			title: "Public invalid method should return 405",
			req: http.Request{
				Method: "POST",
				URL: &url.URL{
					Path: "mock-app/public-success-get",
				},
			},
			expStatusCode: 405,
			resp:          &MockResponseWriter{},
			expResp:       "POST not allowed on mock-app/public-success-get",
			apps:          []App{&MockApp{}},
		},
		{
			title: "Private route with succeeded auth",
			req: http.Request{
				Method: "GET",
				URL: &url.URL{
					Path: "mock-app/private-success-get",
				},
			},
			expStatusCode: 200,
			resp:          &MockResponseWriter{},
			expResp:       "Success",
			apps:          []App{&MockApp{}},
		},
		{
			title: "Private route with failed auth",
			req: http.Request{
				Method: "GET",
				URL: &url.URL{
					Path: "mock-app/private-failure-get",
				},
			},
			expStatusCode: 401,
			resp:          &MockResponseWriter{},
			expResp:       "{\"Msg\": \"Auth failed, please try again\"}",
			apps:          []App{&MockApp{}},
		},
		{
			title: "Private route with failed auth",
			req: http.Request{
				Method: "GET",
				URL: &url.URL{
					Path: "mock-app/private-success-with-no-uid",
				},
			},
			expStatusCode: 401,
			resp:          &MockResponseWriter{},
			expResp:       "{\"Msg\": \"Auth failed, please try again\"}",
			apps:          []App{&MockApp{}},
		},
		{
			title: "Successful GET on mock-app/models",
			req: http.Request{
				Method: "GET",
				URL: &url.URL{
					Path: "mock-app/models",
				},
			},
			expStatusCode: 200,
			resp:          &MockResponseWriter{},
			expResp:       "Success GET",
			apps:          []App{&MockApp{}},
		},
		{
			title: "Successful POST on mock-app/models",
			req: http.Request{
				Method: "POST",
				URL: &url.URL{
					Path: "mock-app/models",
				},
			},
			expStatusCode: 200,
			resp:          &MockResponseWriter{},
			expResp:       "Success POST",
			apps:          []App{&MockApp{}},
		},
		{
			title: "Successful default GET request on mock-app/get",
			req: http.Request{
				Method: "GET",
				URL: &url.URL{
					Path: "mock-app/get",
				},
			},
			expStatusCode: 200,
			resp:          &MockResponseWriter{},
			expResp:       "Success GET",
			apps:          []App{&MockApp{}},
		},
		{
			title: "Successful GET request on mock-app/",
			req: http.Request{
				Method: "GET",
				URL: &url.URL{
					Path: "mock-app/",
				},
			},
			expStatusCode: 200,
			resp:          &MockResponseWriter{},
			expResp:       "Success GET",
			apps:          []App{&MockApp{}},
		},
		{
			title: "Successful default POST request on mock-app/",
			req: http.Request{
				Method: "POST",
				URL: &url.URL{
					Path: "mock-app/",
				},
			},
			expStatusCode: 200,
			resp:          &MockResponseWriter{},
			expResp:       "Success POST",
			apps:          []App{&MockApp{}},
		},
	}

	for _, input := range inputs {

		s := NewService(&Config{}, &input.apps)

		t.Run(input.title, func(t *testing.T) {
			s.ServeHTTP(input.resp, &input.req)
			if mockResp, works := input.resp.(*MockResponseWriter); works {
				if !mockResp.GotExpResp(input.expResp) {
					t.Errorf("expected %s got %s", input.expResp, mockResp.dataWritten)
				}
				if input.expStatusCode != mockResp.statusCode {
					t.Errorf("expected %d got %d", input.expStatusCode, mockResp.statusCode)
				}
			} else {
				t.Error("resp is not MockResponseWriter")
			}
		})
	}

}
