package lib

type Auth struct {
	IsAuthenticated bool
	Payload         any
}

type AuthValidator interface {
	Validate(*Request) Auth
}

type AuthValidatorCallback func(*Service) AuthValidator

type MockAuthValidator struct {
	auth Auth
}

func (v *MockAuthValidator) Validate(req *Request) Auth {
	return v.auth
}

func NewMockAuthValidator(auth Auth) AuthValidatorCallback {
	return func(service *Service) AuthValidator {
		return &MockAuthValidator{
			auth: auth,
		}
	}
}
