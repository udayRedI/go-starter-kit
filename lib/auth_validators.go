package lib

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/getsentry/sentry-go"
)

type AuthValidator interface {
	Validate(*Request) (bool, *string)
}

type RouteServerValidator struct {
	service *Service
}

type AuthValidatorCallback func(*Service) AuthValidator

func NewRouteServerValidator() AuthValidatorCallback {
	return func(service *Service) AuthValidator {
		return &RouteServerValidator{
			service: service,
		}
	}
}

func (rv *RouteServerValidator) Validate(req *Request) (bool, *string) {

	auth := req.GetHeaderVal("Authorization")
	if auth == nil {
		log.Printf("%s Auth is nil", req.ID)
		return false, nil
	}

	if *auth != rv.service.GetAuthToken() {
		return false, nil
	}

	return true, nil

}

type HttpJwtValidator struct {
	jwtValidator *jwtValidator
}

func NewHttpJwtValidator() AuthValidatorCallback {
	return func(service *Service) AuthValidator {
		return &HttpJwtValidator{
			jwtValidator: &jwtValidator{
				service: service,
			},
		}
	}
}

func (v *HttpJwtValidator) Validate(req *Request) (bool, *string) {

	auth := req.GetHeaderVal("Authorization")
	if auth == nil {
		log.Printf("%s Auth not found in header", req.ID)
		return false, nil
	}

	authToken := strings.ReplaceAll(*auth, "Token ", "")

	return v.jwtValidator.Validate(req, authToken)
}

type jwtValidator struct {
	service *Service
}

func (v *jwtValidator) Validate(req *Request, authToken string) (bool, *string) {

	dbFinish, _ := CreateSpan(&req.SentryContext, "JWT verification")
	defer dbFinish()

	httpReq, reqErr := http.NewRequest("GET", v.service.GetValidationUrl(), nil)
	if reqErr != nil {
		log.Println(fmt.Sprintf("%s ERROR: HTTPJwtValidator: Failed to create request with %s: %s", req.ID, v.service.GetValidationUrl(), reqErr))
		return false, nil
	}

	httpReq.Header.Set("Authorization", authToken)

	client := &http.Client{}
	resp, clientErr := client.Do(httpReq)

	if clientErr != nil {
		log.Println(fmt.Sprintf("%s ERROR: HTTPJwtValidator: Failed to run request %s: %s", req.ID, v.service.GetValidationUrl(), clientErr))
		return false, nil
	}

	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		if hub := sentry.GetHubFromContext(req.SentryContext); hub != nil {
			hub.Scope().SetTag("login-type", "JWT")
		}
		log.Println(fmt.Sprintf("%s ERROR: HTTPJwtValidator: Failed  %s with statusCode %d", req.ID, v.service.GetValidationUrl(), resp.StatusCode))

		oktaUid, uidErr := v.getUIDFromJWT(authToken, req)

		if uidErr != nil {
			log.Println(fmt.Sprintf("%s ERROR: HTTPJwtValidator: Failed with %s when decoding authToken %v", req.ID, authToken, uidErr))
			return false, nil
		}
		return true, &oktaUid
	} else {
		fmt.Printf("authtoken: %v and %v", authToken, v.service.StressTestAllowed())
		if v.service.StressTestAllowed() && len(authToken) <= 70 {
			// Assuming that authtoken is uid
			// Should never work on prod
			if hub := sentry.GetHubFromContext(req.SentryContext); hub != nil {
				hub.Scope().SetTag("login-type", "UID")
			}
			print("Rerturning true")
			return true, &authToken
		}
		errorMessage := fmt.Sprintf("Response failed with status %d", resp.StatusCode)
		return false, &errorMessage
	}

}

func (v *jwtValidator) getUIDFromJWT(authToken string, req *Request) (string, error) {

	tokenString := authToken

	// Parse the JWT token without verifying the signature
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return "", fmt.Errorf("%s: failed to parse JWT: %v", req.ID, err)
	}

	// Extract the UID from the token claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("%s: invalid token claims", req.ID)
	}
	uid, ok := claims["uid"].(string)
	if !ok {
		return "", fmt.Errorf("%s: invalid UID in token claims", req.ID)
	}

	return uid, nil
}

type MockAuthValidator struct {
	validated bool
	uid       *string
}

func (v *MockAuthValidator) Validate(req *Request) (bool, *string) {
	return v.validated, v.uid
}

func NewMockAuthValidator(validated bool, errStr *string) AuthValidatorCallback {
	return func(service *Service) AuthValidator {
		return &MockAuthValidator{
			validated: validated,
			uid:       errStr,
		}
	}
}
