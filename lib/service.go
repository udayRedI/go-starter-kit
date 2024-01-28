package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/getsentry/sentry-go"
)

type Service struct {
	config        Config
	Server        http.Server
	apps          map[string]App
	routes        map[string]HttpRoute
	queueHandlers map[string]QueueRoute

	SqsManager   ISqsManager
	CacheManager ICacheManager
}

func NewService(config *Config, definedApps *[]App) *Service {

	s := &Service{
		config:        *config,
		apps:          make(map[string]App),
		routes:        make(map[string]HttpRoute),
		queueHandlers: make(map[string]QueueRoute),
	}

	s.createRoutes(definedApps)

	// Always keep this separate as there's guarantee that all apps are recognised by service.
	for _, app := range *definedApps {
		app.Init(s)
	}

	return s
}

func (s *Service) createRoutes(definedApps *[]App) {
	for _, app := range *definedApps {
		appTitle := app.Title()
		if _, dupFound := s.apps[appTitle]; dupFound {
			errTitle := fmt.Sprintf("%s app already defined, please choose a different name", appTitle)
			CheckFatal(errors.New(errTitle), errTitle)
		}
		s.apps[appTitle] = app
		s.routes[appTitle] = make(HttpRoute)
		s.queueHandlers[appTitle] = make(QueueRoute)
		for _, route := range app.Routes() {
			route.Validate()
			if _, routeFound := s.routes[appTitle][route.Action]; !routeFound {
				s.routes[appTitle][route.Action] = make(map[HttpMethod]HttpAction)
			}
			if _, dupFound := s.routes[appTitle][route.Action][route.Method]; dupFound {
				errTitle := fmt.Sprintf("route re-initialization not allowed for %s action %s method %s", appTitle, route.Action, route.Method)
				CheckFatal(errors.New(errTitle), errTitle)
			}
			s.routes[appTitle][route.Action][route.Method] = route
		}
		for queueRefName, handler := range app.QueueHandlers() {
			queueName, found := s.config.Queues[queueRefName]
			if found == false {
				CheckFatal(errors.New(fmt.Sprintf("%s queue-ref not found in config, please check your config and try again", queueRefName)), "queue listen failed")
			}
			s.queueHandlers[appTitle][queueName] = handler
		}
	}
}

func (s *Service) Init() string {

	s.Server = http.Server{}

	s.Server.Handler = s
	startPort := ":" + s.config.Port
	s.Server.Addr = startPort

	sqsManager, sqsErr := NewSqsManager(s.config.ENV)
	if sqsErr != nil {
		CheckFatal(sqsErr, "SQS initialization failed")
	}

	s.SqsManager = sqsManager

	s.CacheManager = NewRedisManager(&s.config.RedisCreds)

	for _, queues := range s.queueHandlers {
		for queueName, handler := range queues {
			handleErr := sqsManager.HandleQueue(&queueName, handler)
			if handleErr != nil {
				CheckFatal(handleErr, "SQS Handle failed")
			}
		}
	}

	return startPort
}

func (s *Service) GetAppByTitle(title string) App {
	// Only to be used in INIT function as it should fatal if app not found
	app, found := s.apps[title]
	if !found {
		errTxt := fmt.Sprintf("%s not found", title)
		CheckFatal(errors.New(errTxt), errTxt)
	}

	return app
}

func (s *Service) handleAuthResp(req *Request, validators *[]AuthValidatorCallback, onSuccess func(*Request) *Response, onFailure func(*Request) *Response) *Response {

	if validators != nil && len(*validators) > 0 { // Auth check
		for _, validator := range *validators {
			if _isAuthenticated, user := validator(s).Validate(req); _isAuthenticated && user != nil {
				req.isAuthenticated = _isAuthenticated
				req.UserId = *user
				return onSuccess(req)
			} else {
				return onFailure(req)
			}
		}
	}
	req.isAuthenticated = false
	return onSuccess(req)
}

func (s *Service) ServeHTTP(w http.ResponseWriter, httpReq *http.Request) {

	if httpReq.Method != "OPTIONS" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}

	// Set access control headers
	w.Header().Set("Vary", "Accept-Encoding, Authorization, Origin")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if httpReq.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Accept-Encoding, Accept-Language, Access-Control-Allow-Headers, Access-Control-Allow-Methods, Access-Control-Allow-Origin, Access-Control-Max-Age, Access-Control-Request-Headers, Access-Control-Request-Method, Authorization, Origin, Cache-Control, Connection, Content-Type, Content-Encoding, Content-Length, Sec-Fetch-Dest, Sec-Fetch-Mode, Sec-Fetch-Site, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST")
	}

	// No body required if OPTIONS request
	if httpReq.Method == "OPTIONS" {
		fmt.Fprintf(w, "")
		return
	}

	appName, action := decodeURI(httpReq)

	ctx := httpReq.Context()

	var resp *Response

	req := &Request{
		AppTitle:      appName,
		Path:          httpReq.URL.Path,
		Method:        httpReq.Method,
		Header:        httpReq.Header,
		ID:            GenerateRandomUUID(),
		SentryContext: ctx,
		Query:         httpReq.URL.Query(),
	}

	defer Handlepanic(fmt.Sprintf("%s: API (%s) crashed", req.ID, req.Path))

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.Scope().SetTag("trace-id", req.ID)
	}

	returnError := func(errStr string, resp *Response) {
		CaptureSentryException(errStr)
		s.returnResp(w, resp, req)
	}

	methodMap, foundRoute := s.routes[appName][action]
	if !foundRoute {
		returnError(fmt.Sprintf("Invalid route %s encountered in app %s", action, appName), NotFoundResponse())
		return
	}

	var httpAction HttpAction

	if action, actionFound := methodMap[HttpMethod(req.Method)]; actionFound {
		httpAction = action
	} else {

		returnError(fmt.Sprintf("%s not allowed on %s", req.Method, httpReq.URL.Path), &Response{
			Status: http.StatusMethodNotAllowed,
			Body:   fmt.Sprintf("%s not allowed on %s", req.Method, httpReq.URL.Path),
		})
		return
	}

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.Scope().SetTag("RequestType", "HTTP")
	}

	resp = s.handleAuthResp(req, &httpAction.AuthValidators, func(r *Request) *Response {
		return httpAction.Handler(req)
	}, func(r *Request) *Response {
		return AuthFailedResponse()
	})

	s.returnResp(w, resp, req)

}

func (s *Service) prepareResp(w http.ResponseWriter, resp *Response, req *Request) *[]byte {
	// Prepare HTTP response
	httpRespStr, respIsStr := resp.Body.(string)
	httpRespBytes := []byte(httpRespStr)
	if !respIsStr {
		var respIsBytes bool
		httpRespBytes, respIsBytes = resp.Body.([]byte)
		if !respIsBytes {
			httpRespJSONBytes, err := json.Marshal(resp.Body)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				CaptureSentryException(fmt.Sprintf("%s Error encountered during json.Marshal for body %v, error %s", req.ID, resp.Body, err))
				fmt.Fprintf(w, "{\"Msg\": \"something went wrong on our side\"}")
				return nil
			}
			httpRespBytes = httpRespJSONBytes
		}
	}
	return &httpRespBytes
}

func (s *Service) returnResp(w http.ResponseWriter, resp *Response, req *Request) {

	bytesResp := s.prepareResp(w, resp, req)
	if resp.Status != 0 {
		w.WriteHeader(resp.Status)
	}

	if bytesResp == nil {
		log.Printf("%s bytesResp is nil", req.ID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err := w.Write(*bytesResp)
	if err != nil {
		log.Printf("%s Error during w.Write with %s", req.ID, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "{\"Msg\": \"something went wrong on our side\"}")
	}
}
