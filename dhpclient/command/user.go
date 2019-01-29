package command

import (
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

type UserCommand struct {
	Revision          string
	Version           string
	VersionPrerelease string
	Ui                cli.Ui
}

func (uc *UserCommand) Help() string {
	helpText := `
Usage: dhpclient user [options]
  Performs calls to the DHP user management  assembly
Options:
  -action=profile   The action to perform, defaults to getting the user profile (profile)
                    Available actions: [profile, prefs]
  -user=            The user UUID to use in actions
  -token=           A user access token
  -key=             The key to set
  -val=             The value of the key. If not set key will be deleted
	`
	return strings.TrimSpace(helpText)
}

func (uc *UserCommand) Synopsis() string {
	return "Performs user management API requests"
}

func (uc *UserCommand) Run(args []string) int {
	var apiEndpoint string
	var queryParams string
	var method string
	var header *http.Header = &http.Header{}

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

	cmdFlags := flag.NewFlagSet("user", flag.ContinueOnError)
	cmdFlags.Usage = func() { uc.Ui.Output(uc.Help()) }
	action := cmdFlags.String("action", "profile", "Type of call. Defaults to profile")
	userId := cmdFlags.String("user", "", "The user id to use in the call")
	token := cmdFlags.String("token", "", "A user access token")
	key := cmdFlags.String("key", "", "The key to set")
	val := cmdFlags.String("val", "", "The value of the key")
	if err := cmdFlags.Parse(args); err != nil {
		log.Error(err)
		return 1
	}

	var body = []byte{}

	switch *action {
	case "profile":
		if *userId == "" || *token == "" {
			log.Error("user and token required to get profile")
			return 1
		}
		method = "GET"
		apiEndpoint = "/usermanagement/users/" + *userId + "/profile"
		queryParams = "applicationName=" + config.DhpApplicationName
	case "prefs":
		if *userId == "" || *token == "" {
			log.Error("userid, access token and key required to set prefs")
			return 1
		}
		// Fetch the profile first
		method = "GET"
		apiEndpoint = "/usermanagement/users/" + *userId + "/profile"
		queryParams = "applicationName=" + config.DhpApplicationName
		header.Add("accessToken", *token)
		response := c.SendRestRequest(method, apiEndpoint, queryParams, header, body)
		if response.StatusCode != 200 {
			log.Error("profile not found")
			return 1
		}
		jsonParsed, _ := gabs.ParseJSON([]byte(response.Body))
		// Tweak profile here
		profile := jsonParsed.Path("exchange.user.profile")
		if *key == "" {
			// Dump the profile if no key is provided
			fmt.Println(profile.StringIndent("", "  "))
			return 0
		}

		if *val == "" {
			profile.Delete("preferences", *key)
		} else {
			profile.Set(*val, "preferences", *key)
		}
		body = profile.Bytes()
		method = "PUT"
	default:
		return 1
	}
	var response client.Response
	if *token == "" {
		response = c.SendSignedRequest(method, apiEndpoint, queryParams, header, body)
	} else {
		header.Add("AccessToken", *token)
		response = c.SendRestRequest(method, apiEndpoint, queryParams, header, body)
	}
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
