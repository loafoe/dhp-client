package command

import (
	"encoding/json"
	"flag"
	"fmt"
	"git.aemian.com/dhp/client"
	"github.com/Jeffail/gabs/v2"
	"github.com/loafoe/cfutil"
	"github.com/mitchellh/cli"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

type CustomerCare struct {
	Revision          string
	Version           string
	VersionPrerelease string
	Ui                cli.Ui
}

type ResetPasswordWithCodeRequest struct {
	Code            string `json:"code"`
	NewPassword     string `json:"newPassword"`
	ConfirmPassword string `json:"confirmPassword"`
}

func (cc *CustomerCare) Help() string {
	helpText := `
Usage: dhpclient cc [options]
  Performs calls to the CustomerCare
Options:
  -action=profile   The action to perform, defaults to getting the user profile (profile)
                    Available actions: profile, recovery, smsreset, logout
  -username=        The username (email) to lookup
  -code=            SMS code (smsreset)
  -version=         API version to use
  -password=        New password (smsreset)
	`
	return strings.TrimSpace(helpText)
}

func (cc *CustomerCare) Synopsis() string {
	return "Performs customer care API requests"
}

func (cc *CustomerCare) Run(args []string) int {
	var apiEndpoint string
	var queryParams string
	var method string
	var header *http.Header = &http.Header{}
	var body []byte

	config := client.ApiClientConfig{
		ApiBaseUrl:         cfutil.Getenv("DHP_AUTH_URL"),
		DhpApplicationName: cfutil.Getenv("DHP_APPLICATION_NAME"),
		SigningKey:         cfutil.Getenv("DHP_CUSTOMERCARE_SIGNING_KEY"),
		SigningSecret:      cfutil.Getenv("DHP_CUSTOMERCARE_SIGNING_SECRET"),
	}
	c, err := client.NewClient(config)
	if err != nil {
		log.Error(err)
		return 1
	}
	cmdFlags := flag.NewFlagSet("cc", flag.ContinueOnError)
	cmdFlags.Usage = func() { cc.Ui.Output(cc.Help()) }
	action := cmdFlags.String("action", "profile", "Type of call. Defaults to profile")
	username := cmdFlags.String("username", "", "The username (email) to lookup")
	password := cmdFlags.String("password", "", "The password")
	version := cmdFlags.String("version", "2", "The API version to use. Default is 2")
	date := cmdFlags.String("date", "", "The SignedDate value to use")
	code := cmdFlags.String("code", "", "An SMS code")
	if err := cmdFlags.Parse(args); err != nil {
		log.Error(err)
		return 1
	}
	switch *action {
	case "profile":
		if *username == "" {
			log.Error("username required")
			return 1
		}
		method = "POST"
		apiEndpoint = "/usermanagement/users/profile"
		queryParams = "applicationName=" + config.DhpApplicationName
		body, _ = json.Marshal(&AuthRequest{
			LoginId: *username,
		})
	case "logout":
		if *username == "" {
			log.Error("username required")
			return 1
		}
		method = "POST"
		apiEndpoint = "/authentication/users/logout"
		queryParams = "applicationName=" + config.DhpApplicationName
		body, _ = json.Marshal(&AuthRequest{
			LoginId: *username,
		})
	case "recovery":
		if *username == "" {
			log.Error("username required")
			return 1
		}
		method = "POST"
		apiEndpoint = "/authentication/credential/sendPasswordRecoveryCode"
		queryParams = "applicationName=" + config.DhpApplicationName
		header.Set("Api-Version", "1")
		body, _ = json.Marshal(&AuthRequest{
			LoginId: *username,
		})
	case "smsreset":
		if *code == "" {
			log.Error("SMS code required")
			return 1
		}
		if *password == "" {
			log.Error("New password required")
			return 1
		}
		method = "POST"
		apiEndpoint = "/authentication/credential/changePasswordWithSMSCode"
		queryParams = "applicationName=" + config.DhpApplicationName
		body, _ = json.Marshal(&ResetPasswordWithCodeRequest{
			NewPassword:     *password,
			ConfirmPassword: *password,
			Code:            *code,
		})
	default:
		log.Print("Unknown action: %s\n", *action)
		return 1
	}
	if *date != "" {
		header.Set("SignedDate", *date)
	}
	if *version != "" {
		header.Set("Api-Version", *version)
	} else {
		header.Set("Api-Version", "1")
	}
	response := c.SendSignedRequest(method, apiEndpoint, queryParams, header, body)
	if response.StatusCode < 200 {
		log.Print(string(response.Body))
		for _, e := range response.Errors {
			log.Print(e)
		}
	}
	jsonParsed, err := gabs.ParseJSON([]byte(response.Body))
	if err != nil {
		log.Print(err)
		return 1
	}
	fmt.Println(jsonParsed.StringIndent("", "  "))

	return 0
}
