package lib

type OktaConfig struct {
	Api           string `json:"OKTA_API"`
	Token         string `json:"OKTA_API_TOKEN"`
	Issuer        string `json:"OKTA_ISSUER"`
	RetailGroupId string `json:"OKTA_RETAIL_GROUP_ID"`
}

func (OktaConfig *OktaConfig) IsValid() bool {
	return len(OktaConfig.Api) > 0 && len(OktaConfig.Token) > 0 && len(OktaConfig.Issuer) > 0 && len(OktaConfig.RetailGroupId) > 0
}

type Config struct {
	Port        string `json:"Port"`
	Name        string `json:"Name"`
	Version     string `json:"Version"`
	VersionDate string `json:"VersionDate"`

	ENV             string  `json:"ENV"` // DEV, QA, PROD
	DSN             string  `json:"DSN"` // Sentry DSN URL
	TraceSampleRate float64 `json:"TraceSampleRate"`

	ESHost          string `json:"ESHost"`
	ESPort          string `json:"ESPort"`
	ESUser          string `json:"ESUser"`
	ESPass          string `json:"ESPass"`
	AuthToken       string `json:"AuthToken"`
	AllowStressTest bool   `json:"AllowStressTest"`

	JwtValidationUrl string `json:"JwtValidationUrl"`
	DroneApiUrl      string `json:"DroneApiUrl"`
	SegmentWriteKey  string `json:"SegmentWriteKey"`

	NotificationApiUrl string `json:"NotificationApiUrl"`

	AWSSecrets map[string]string `json:"AWSSecrets"`

	Queues map[string]string `json:"Queues"`

	DbUrl      string     `json:"DbUrl"`
	RedisCreds CacheCreds `json:"Redis"`
}

func (config *Config) IsValid() bool {
	return true
}
