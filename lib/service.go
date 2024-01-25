package lib

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
)

type Service struct {
	config        Config
	Server        http.Server
	apps          map[string]App
	routes        map[string]HttpRoute
	queueHandlers map[string]QueueRoute
	sseHandlers   map[string]SseRoute

	SqsManager       ISqsManager
	CacheManager     ICacheManager
	AnalyticsManager IAnalyticsManager
}

func NewService(config *Config, definedApps *[]App) *Service {

	s := &Service{
		config:        *config,
		apps:          make(map[string]App),
		routes:        make(map[string]HttpRoute),
		queueHandlers: make(map[string]QueueRoute),
		sseHandlers:   make(map[string]SseRoute),
	}

	for _, app := range *definedApps {
		appTitle := app.Title()
		s.apps[appTitle] = app
		s.routes[appTitle] = make(HttpRoute)
		s.queueHandlers[appTitle] = make(QueueRoute)
		s.sseHandlers[appTitle] = make(SseRoute)
		for routeName, route := range app.Routes() {
			s.routes[appTitle][routeName] = route
		}
		for queueRefName, handler := range app.QueueHandlers() {
			queueName, found := s.config.Queues[queueRefName]
			if found == false {
				CheckFatal(errors.New(fmt.Sprintf("%s queue-ref not found in config, please check your config and try again", queueRefName)), "queue listen failed")
			}
			s.queueHandlers[appTitle][queueName] = handler
		}
		for sseRoute, sseHandler := range app.SseHandlers() {
			s.sseHandlers[appTitle][sseRoute] = sseHandler
		}
	}

	// Always keep this separate as there's guarantee that all apps are recognised by service.
	for _, app := range *definedApps {
		app.Init(s)
	}

	return s
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
	s.AnalyticsManager = NewAnalyticsManager(s.config.SegmentWriteKey)

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

	appName, action, recordID := decodeURI(httpReq)

	// TODO: Move to a factory
	freeTextQuery, filter, sort, page, pageSize, fields, context, uiReady, other, scrollId := decodeQueryParams(httpReq.URL.Query())

	ctx := httpReq.Context()

	var resp *Response

	req := &Request{
		AppTitle:         appName,
		Action:           action,
		Method:           httpReq.Method,
		RecordID:         recordID,
		FreeTextQuery:    freeTextQuery,
		Filter:           filter,
		Sort:             sort,
		Page:             page,
		PageSize:         pageSize,
		Fields:           fields,
		Context:          context,
		UIReady:          uiReady,
		OtherQueryParams: other,
		Body:             httpReq.Body,
		scrollId:         scrollId,
		Header:           httpReq.Header,
		ID:               GenerateRandomUUID(),
		SentryContext:    ctx,
	}

	defer Handlepanic(fmt.Sprintf("%s: API (%s) crashed", req.ID, action))

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.Scope().SetTag("trace-id", req.ID)
	}

	actionRoute, foundRoute := s.routes[appName][action]
	if foundRoute {
		var isAuthenticated bool = true // default true as there could be public routes and things are handled in actionRoute.AuthValidator
		var userIdStr string = ""
		if actionRoute.AuthValidator != nil { // Auth check
			validator := actionRoute.AuthValidator(s)
			if validator == nil {
				isAuthenticated = false
			} else {
				_isAuthenticated, _userPtr := validator.Validate(req)
				isAuthenticated = _isAuthenticated
				if _userPtr != nil {
					userIdStr = *_userPtr
				}
			}
		}
		req.isAuthenticated = isAuthenticated
		req.UserId = userIdStr
		if hub := sentry.GetHubFromContext(ctx); hub != nil {
			hub.Scope().SetTag("RequestType", "HTTP")
		}
		if req.isAuthenticated != true {
			resp = AuthFailedResponse()
		} else {
			handler := actionRoute.Handler
			resp = handler(req)
		}
	} else {
		if sseAction, foundSseHandler := s.sseHandlers[appName][action]; foundSseHandler {
			if hub := sentry.GetHubFromContext(ctx); hub != nil {
				hub.Scope().SetTag("RequestType", "SSE")
			}

			if flusher, ok := w.(http.Flusher); ok {

				if sseAction.AuthValidator != nil {
					validator := sseAction.AuthValidator(s)
					if validator == nil {
						req.isAuthenticated = false
					} else {
						_isAuthenticated, _userPtr := validator.Validate(req)
						req.isAuthenticated = _isAuthenticated
						if _userPtr != nil {
							req.UserId = *_userPtr
						} else {
							CaptureSentryException(fmt.Sprintf("%s Somehow user is authenticated but user-id is undefined %s", req.ID, req.UserId))
						}
					}
				} else {
					req.isAuthenticated = true
				}
				if !req.isAuthenticated {
					resp = AuthFailedResponse()
				} else {
					successChan, errChan, handleClose := sseAction.Handler(req)
					w.Header().Set("Content-Type", "text/event-stream")
					w.Header().Set("Cache-Control", "no-cache")
					w.Header().Set("Connection", "keep-alive")
					if successChan == nil || errChan == nil {
						sseHandleErr := errors.New(fmt.Sprintf("SuccessChan(%v) errChan(%v) cannot be nil ", successChan, errChan))
						CaptureSentryException(sseHandleErr.Error())
						resp = ClientErrorResponse(sseHandleErr)
					} else {
					mainOut:
						for {
							select {
							case <-time.After(55 * time.Second):
								fmt.Fprintf(w, "data: ping\n\n")
								flusher.Flush()
							case successResp := <-*successChan:
								fmt.Fprintf(w, "data: %s\n\n", successResp)
								flusher.Flush()
							case chanErr := <-*errChan:
								log.Printf("%s: INFO: Closing connection %s because of error %s", req.ID, req.UserId, chanErr)
								handleClose(req)
								break mainOut
							case <-req.SentryContext.Done():
								log.Println(fmt.Sprintf("%s: INFO: Connection closed for UID: %s", req.ID, req.UserId))
								handleClose(req)
								break mainOut
							}
						}
					}
				}

			} else {
				CaptureSentryException(fmt.Sprintf("Route %s doesnt have http.Flusher for app %s", action, appName))
				resp = NotFoundResponse()
			}
		} else {
			CaptureSentryException(fmt.Sprintf("Invalid route %s encountered in app %s", action, appName))
			resp = NotFoundResponse()
		}
	}

	// Prepare HTTP response
	httpRespStr, foundRoute := resp.Body.(string)
	httpRespBytes := []byte(httpRespStr)
	if !foundRoute {
		var okBytes bool
		httpRespBytes, okBytes = resp.Body.([]byte)
		foundRoute = okBytes
	}
	if !foundRoute {
		httpRespJSONBytes, err := json.Marshal(resp.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			CaptureSentryException(fmt.Sprintf("Error encountered during json.Marshal for body %v, error %s", resp.Body, err))
			fmt.Fprintf(w, "{\"Msg\": \"something went wrong on our side\"}")
			return
		}
		httpRespBytes = httpRespJSONBytes
	}

	// Send response
	if resp.Status != 0 && resp.Status != http.StatusOK {
		w.WriteHeader(resp.Status)
	}
	s.compressResponse(w, httpReq, &httpRespBytes)
	_, err := w.Write(httpRespBytes)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "{\"Msg\": \"something went wrong on our side\"}")
	}
}

// compressResponse gzips large responses for better load performance.
func (s *Service) compressResponse(w http.ResponseWriter, httpReq *http.Request, resp *[]byte) {
	// Small responses needn't be compressed
	if len(*resp) <= 1024 {
		return
	}

	// Ensure client accepts gzip
	if !strings.Contains(httpReq.Header.Get("Accept-Encoding"), "gzip") {
		return
	}

	// Skip those with missing content type
	contentType := w.Header().Get("Content-Type")
	if contentType == "" {
		return
	}

	// Don't re-compress
	if strings.HasPrefix(contentType, "image/") ||
		strings.HasPrefix(contentType, "audio/") ||
		strings.HasPrefix(contentType, "video/") ||
		strings.HasPrefix(contentType, "zip/") ||
		strings.HasPrefix(contentType, "zip2/") {
		return
	}

	// Compress response
	compressedResp := new(bytes.Buffer)
	compressor := gzip.NewWriter(compressedResp)

	// Send uncompressed response if there is a compression error
	if _, err := compressor.Write(*resp); err != nil {
		compressor.Close()
		return
	}
	if err := compressor.Close(); err != nil {
		return
	}

	// Compression successful, update response and encoding header
	c := compressedResp.Bytes()
	*resp = c
	w.Header().Add("Content-Encoding", "gzip")
}
