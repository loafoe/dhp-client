package command

import (
	_ "encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"git.aemian.com/dhp/client"
	log "github.com/Sirupsen/logrus"
	"github.com/jeffail/gabs"
	"github.com/loafoe/cfutil"
	"github.com/mitchellh/cli"
)

type RequestCommand struct {
	Revision          string
	Version           string
	VersionPrerelease string
	Ui                cli.Ui
}

func (uc *RequestCommand) Help() string {
	helpText := `
Usage: dhpclient request [options]
  Request
Options:
  -service=authorize|subscription     Authorize or Subscription request. Defaults to subscription
  -method=GET|POST|PUT|DELETE|PATCH   Request method 
  -body=stringBody                    Optional body of request
  -version=                           Optional set Api-Version header
  -endpoint=/some/path                The endpoint of the request
  -headers="..."                      Headers to add. Separate with ;
	`
	return strings.TrimSpace(helpText)
}

func (uc *RequestCommand) Synopsis() string {
	return "Request command"
}

func (uc *RequestCommand) Run(args []string) int {
	var header *http.Header = &http.Header{}
	cmdFlags := flag.NewFlagSet("test", flag.ContinueOnError)
	cmdFlags.Usage = func() { uc.Ui.Output(uc.Help()) }
	method := cmdFlags.String("method", "GET", "Type of request method. Defaults to GET")
	endpoint := cmdFlags.String("endpoint", "", "The endpoint")
	version := cmdFlags.String("version", "", "The API version to use")
	body := cmdFlags.String("body", "", "Optional JSON body of request")
	headers := cmdFlags.String("headers", "", "Headers to add to request. Separate with ;")
	service := cmdFlags.String("service", "subscription", "IAM or IDM request")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}
	baseURL := cfutil.Getenv("DHP_SUBSCRIPTION_SERVICE_URL")
	signingKey := cfutil.Getenv("DHP_SUBSCRIPTION_SIGNING_KEY")
	signingSecret := cfutil.Getenv("DHP_SUBSCRIPTION_SIGNING_SECRET")

	if *service == "authorize" {
		baseURL = cfutil.Getenv("DHP_AUTH_URL")
		signingKey = cfutil.Getenv("DHP_CUSTOMERCARE_SIGNING_KEY")
		signingSecret = cfutil.Getenv("DHP_CUSTOMERCARE_SIGNING_SECRET")
	}

	config := client.ApiClientConfig{
		ApiBaseUrl:         baseURL,
		DhpApplicationName: cfutil.Getenv("DHP_APPLICATION_NAME"),
		SigningKey:         signingKey,
		SigningSecret:      signingSecret,
	}
	c, err := client.NewClient(config)
	if err != nil {
		log.Error(err)
		return 1
	}
	if *endpoint == "" {
		fmt.Println("endpoint is required")
		return 2
	}

	// Add headers
	splittedHeaders := strings.Split(*headers, ";")
	for _, h := range splittedHeaders {
		kv := strings.Split(h, ":")
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		header.Set(key, val)
	}
	apiEndpoint := *endpoint
	queryParams := ""

	if *version != "" {
		header.Set("Api-Version", *version)
	}
	response := c.SendSignedRequest(*method, apiEndpoint, queryParams, header, []byte(*body))
	if response.StatusCode < 200 {
		log.Print(string(response.Body))
		for _, e := range response.Errors {
			log.Print(e)
		}
		return 1
	}
	jsonParsed, err := gabs.ParseJSON([]byte(response.Body))
	if err != nil {
		log.Print("Error decoding body:")
		log.Print(err)
		log.Print("RAW RESPONSE:")
		log.Print(string(response.Body))
		return 1
	}
	fmt.Println(jsonParsed.StringIndent("", "  "))

	return 0
}
