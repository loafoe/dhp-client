package command

import (
	"encoding/json"
	_ "encoding/json"
	"flag"
	"fmt"
	"git.aemian.com/dhp/client"
	log "github.com/Sirupsen/logrus"
	"github.com/jeffail/gabs"
	"github.com/loafoe/cfutil"
	"github.com/mitchellh/cli"
	"net/http"
	"strings"
)

type AuthRequest struct {
	LoginId  string `json:"loginId,omitempty"`
	Password string `json:"password,omitempty"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type AuthCommand struct {
	Revision          string
	Version           string
	VersionPrerelease string
	Ui                cli.Ui
}

func (au *AuthCommand) Help() string {
	helpText := `
Usage: dhpclient auth [options]
  Performs requests to the DHP authorization endpoints in the user management assembly
Options:
  -action=login               Type of auth action, defaults to login
                              Available actions: login, logout, status, refresh, recover
  -username=  		      The login id of the user
  -password=                  The password
  -secret=                    OPtional refresh secret
  -token=                     An access or refresh token
  -user=                      An UUID of the user"
	`
	return strings.TrimSpace(helpText)
}

func (au *AuthCommand) Synopsis() string {
	return "Performs authorization requests on the management assembly"
}

func (au *AuthCommand) Run(args []string) int {
	var apiEndpoint string
	var queryParams string
	var method string
	var header *http.Header = &http.Header{}
	var body []byte

	// Setup configuration for the apiClient
	config := client.ApiClientConfig{
		ApiBaseUrl:         cfutil.Getenv("DHP_AUTH_URL"),
		DhpApplicationName: cfutil.Getenv("DHP_APPLICATION_NAME"),
		SigningKey:         cfutil.Getenv("DHP_SIGNING_KEY"),
		SigningSecret:      cfutil.Getenv("DHP_SIGNING_SECRET"),
	}
	apiClient, err := client.NewClient(config)
	if err != nil {
		log.Error(err)
		return 1
	}

	// Setup and parse parameters
	cmdFlags := flag.NewFlagSet("auth", flag.ContinueOnError)
	cmdFlags.Usage = func() { au.Ui.Output(au.Help()) }
	action := cmdFlags.String("action", "login", "Type of action. Defaults to login")
	loginId := cmdFlags.String("username", "", "User name")
	password := cmdFlags.String("password", "", "User password")
	token := cmdFlags.String("token", "", "Access or Refresh token")
	userId := cmdFlags.String("user", "", "The user UUID")
	refreshSecret := cmdFlags.String("secret", "", "The refresh secret")
	if err := cmdFlags.Parse(args); err != nil {
		log.Error(err)
		return 1
	}

	switch *action {
	case "login":
		if *loginId == "" || *password == "" {
			log.Error("username and password must be provided")
			return 1
		}
		method = "POST"
		apiEndpoint = "/authentication/login"
		queryParams = "applicationName=" + config.DhpApplicationName
		body, _ = json.Marshal(&AuthRequest{
			LoginId:  *loginId,
			Password: *password,
		})
		if *refreshSecret != "" {
			header.Add("refreshSecret", *refreshSecret)
		}
		header.Add("Api-Version", "2")

	case "logout":
		if *userId == "" || *token == "" {
			log.Error("user and token must be provided")
			return 1
		}
		method = "POST"
		apiEndpoint = "/authentication/users/" + *userId + "/logout"
		queryParams = "applicationName=" + config.DhpApplicationName
		header.Add("Authorization", "Bearer "+*token)
	case "status":
		if *userId == "" || *token == "" {
			log.Error("user and token must be provided")
			return 1
		}
		method = "GET"
		apiEndpoint = "/authentication/users/" + *userId + "/tokenStatus"
		queryParams = "applicationName=" + config.DhpApplicationName
		header.Add("AccessToken", *token)
	case "refresh":
		if *userId == "" || *token == "" {
			log.Error("user and token must be provided")
			return 1
		}
		method = "PUT"
		apiEndpoint = "/authentication/users/" + *userId + "/refreshToken"
		queryParams = "applicationName=" + config.DhpApplicationName
		body, _ = json.Marshal(&RefreshRequest{
			RefreshToken: *token,
		})
	case "recover":
		if *loginId == "" {
			log.Error("username must be provided")
			return 1
		}
		method = "POST"
		apiEndpoint = "/authentication/credential/recoverPassword"
		queryParams = "applicationName=" + config.DhpApplicationName
		body, _ = json.Marshal(&AuthRequest{
			LoginId: *loginId,
		})

	default:
		return 1
	}
	response := apiClient.SendSignedRequest(method, apiEndpoint, queryParams, header, body)
	if response.StatusCode < 200 {
		log.Print(string(response.Body))
		for _, e := range response.Errors {
			log.Print(e)
		}
	}
	jsonParsed, err := gabs.ParseJSON([]byte(response.Body))
	if err != nil {
		log.Error(err)
		return 1
	}
	fmt.Println(jsonParsed.StringIndent("", "  "))
	return 0
}
