package command

import (
	_ "encoding/json"
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

type SubscriptionCommand struct {
	Revision          string
	Version           string
	VersionPrerelease string
	Ui                cli.Ui
}

func (sc *SubscriptionCommand) Help() string {
	helpText := `
Usage: dhpclient subscription [options]
  Performs calls to the DHP subscription assembly
Options:
  -action=tc        Type of call, defaults to Terms & Conditions (tc)
                    Available action: [tc, close]
  -consent=         The consent code. Default=2		    
  -user=  			The user UUID to use in actions
	`
	return strings.TrimSpace(helpText)
}

func (sc *SubscriptionCommand) Synopsis() string {
	return "Performs subscription API requests"
}

func (sc *SubscriptionCommand) Run(args []string) int {
	var apiEndpoint string
	var queryParams string
	var method string
	var header *http.Header = &http.Header{}

	config := client.ApiClientConfig{
		ApiBaseUrl:         cfutil.Getenv("DHP_SUBSCRIPTION_SERVICE_URL"),
		DhpApplicationName: cfutil.Getenv("DHP_APPLICATION_NAME"),
		SigningKey:         cfutil.Getenv("DHP_SUBSCRIPTION_SIGNING_KEY"),
		SigningSecret:      cfutil.Getenv("DHP_SUBSCRIPTION_SIGNING_SECRET"),
		PropositionName:    cfutil.Getenv("DHP_PROPOSITION_NAME"),
	}
	c, err := client.NewClient(config)
	if err != nil {
		log.Error(err)
		return 1
	}

	cmdFlags := flag.NewFlagSet("subscription", flag.ContinueOnError)
	cmdFlags.Usage = func() { sc.Ui.Output(sc.Help()) }
	action := cmdFlags.String("action", "tc", "Type of call. Defaults to terms and conditions")
	userId := cmdFlags.String("user", "", "The user id to use in the call")
	consent := cmdFlags.String("consent", "2", "The consent code")
	token := cmdFlags.String("token", "", "A user access token")
	if err := cmdFlags.Parse(args); err != nil {
		log.Error(err)
		return 1
	}
	body := []byte{}
	switch *action {
	case "close":
		if *userId == "" {
			log.Error("user-id required to get Terms & Conditions")
			return 1
		}

		method = "PUT"
		apiEndpoint = "/subscription/applications/" + config.DhpApplicationName + "/users/" + *userId + "/close"
		body = []byte("{\"deleteDataFlag\":\"Yes\"}")

	case "tc":
		method = "GET"
		apiEndpoint = "/subscription/applications/" + config.DhpApplicationName + "/users/" + *userId + "/termsAndConditions"
		queryParams = "consentCode=" + *consent + "&propositionName=" + config.PropositionName
		if *userId == "" {
			log.Error("user-id required to get Terms & Conditions")
			return 1
		}
		if *token != "" {
			header.Add("accessToken", *token)
		}
	default:
		log.Error("Unknown action ", *action)
		return 1
	}
	header.Add("Api-Version", "1")
	response := c.SendSignedRequest(method, apiEndpoint, queryParams, header, body)
	if response.StatusCode < 200 {
		log.Print(string(response.Body))
		for _, e := range response.Errors {
			log.Print(e)
		}
	}
	jsonParsed, err := gabs.ParseJSON([]byte(response.Body))
	if err == nil {
		fmt.Println(jsonParsed.StringIndent("", "  "))
	} else {
		log.Error(err)
		return 1
	}

	return 0
}
