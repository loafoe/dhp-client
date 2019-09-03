package command

import (
	_ "encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"git.aemian.com/dhp/client"
	"github.com/loafoe/cfutil"
	"github.com/mitchellh/cli"
	log "github.com/sirupsen/logrus"
)

type SignCommand struct {
	Revision          string
	Version           string
	VersionPrerelease string
	Ui                cli.Ui
}

func (sr *SignCommand) Help() string {
	helpText := `
Usage: signer sign [options]
  Generate DHP signature
Options:
  -method=               HTTP method, defaults to GET
  -body=                 The request body or path to fil File which contains request body
  -date= 		 Optionally inject signed date header value
  -path=                 Path of the request e.g. /api/do/something
  -api=                  Optional Api-Version to send
  -params=               Param list
	`
	return strings.TrimSpace(helpText)
}

func (sr *SignCommand) Synopsis() string {
	return "Generate DHP signature"
}

func (sr *SignCommand) Run(args []string) int {
	var header *http.Header = &http.Header{}

	signingKey := cfutil.Getenv("DHP_CUSTOMERCARE_SIGNING_KEY")
	signingSecret := cfutil.Getenv("DHP_CUSTOMERCARE_SIGNING_SECRET")

	// Setup configuration for the apiClient
	config := client.ApiClientConfig{
		ApiBaseUrl:         "http://dummy-host",
		DhpApplicationName: "dummyName",
		SigningKey:         signingKey,
		SigningSecret:      signingSecret,
		Debug:              os.Getenv("DHP_CLIENT_DEBUG") == "true",
	}
	apiClient, err := client.NewClient(config)
	if err != nil {
		log.Error(err)
		return 1
	}

	// Setup and parse parameters
	cmdFlags := flag.NewFlagSet("sign", flag.ContinueOnError)
	cmdFlags.Usage = func() { sr.Ui.Output(sr.Help()) }
	method := cmdFlags.String("method", "GET", "HTTP method, defaults to GET")
	date := cmdFlags.String("date", "", "SignedDate value defaults to currnt timestamp")
	params := cmdFlags.String("params", "", "Parameter list")
	bodyFile := cmdFlags.String("body", "", "The request body or path to file which contains request body")
	path := cmdFlags.String("path", "", "The request path e.g. /api/do/something")
	api := cmdFlags.String("api", "", "Optional Api-version to send in Api-Header")

	if err := cmdFlags.Parse(args); err != nil {
		log.Error(err)
		return 1
	}
	if *path == "" {
		log.Error("path must be provided")
		return 1
	}
	var body []byte

	body, err = ioutil.ReadFile(*bodyFile)
	if err != nil {
		body = []byte(*bodyFile)
	}
	if *date != "" {
		header.Set("SignedDate", *date)
	}
	if *api != "" {
		header.Set("Api-Version", *api)
	}

	apiClient.Sign(time.Now(), header, *path, *params, *method, body)
	log.Info("Key    = ", signingKey)
	log.Info("Secret = ", signingSecret)
	log.Info("Path   = ", *path)
	log.Info("Params = ", *params)
	log.Info("Method = ", *method)
	log.Info("Body:")
	log.Info(string(body))
	log.Info("------------ Signed result -------------")
	log.Info("SignedDate: ", header.Get("SignedDate"))
	log.Info("Authorization: ", header.Get("Authorization"))
	return 0
}
