package client

type ApiClientConfig struct {
	ApiBaseUrl         string
	DhpApplicationName string
	SigningKey         string
	SigningSecret      string
	PropositionName    string
	Debug              bool
}

func (config *ApiClientConfig) Init(apiBaseUrl, dhpApplicationName, signingKey, signingSecret, propositionName string, debug bool) {
	config.ApiBaseUrl = apiBaseUrl
	config.DhpApplicationName = dhpApplicationName
	config.SigningKey = signingKey
	config.SigningSecret = signingSecret
	config.PropositionName = propositionName
	config.Debug = debug
}
