// This package provides an API client for interacting with the HSDP services (DHP)
package client

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jeffail/gabs"
)

var (
	TIME_FORMAT = "2006-01-02T15:04:05.000-0700"
)

// The API client
type ApiClient struct {
	config             ApiClientConfig
	apiSigner          ApiSigner
	apiBaseUrl         string
	dphApplicationName string
}

// NewClient creates a new client. It takes an ApiClientConfig struct
// for configuration
func NewClient(config ApiClientConfig) (*ApiClient, error) {
	client := &ApiClient{}
	error := client.Init(config)
	return client, error
}

// You can call the Init() function when creating a client without
// using the NewClient() call
func (client *ApiClient) Init(config ApiClientConfig) error {
	client.config = config
	client.apiBaseUrl = config.ApiBaseUrl
	if client.apiBaseUrl == "" {
		return errors.New("BaseUrl must be provided")
	}
	if os.Getenv("DHP_CLIENT_DEBUG") == "true" {
		client.config.Debug = true
	}
	client.apiSigner.Init(config.SigningKey, config.SigningSecret, config.Debug)
	client.dphApplicationName = config.DhpApplicationName
	return nil
}

func (client *ApiClient) sendRestRequest(httpMethod string, uri *url.URL, header *http.Header, body []byte) Response {
	if header.Get("Content-Type") == "" {
		header.Set("Content-Type", "application/json")
	}
	if header.Get("Accept") == "" {
		header.Set("Accept", "application/json")
	}

	buf := bytes.NewBuffer(body)
	req, err := http.NewRequest(httpMethod, uri.String(), buf)
	if err != nil {
		log.Error("Request failed: ", err.Error())
		return Response{
			Body: "error",
		}
	}
	for k, _ := range *header {
		req.Header.Set(k, header.Get(k))
	}

	if client.config.Debug {
		dumped, _ := httputil.DumpRequest(req, true)
		log.Info(string(dumped))
	}

	// Fetch Request
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		log.Error("Request failed: ", err)
	}

	if client.config.Debug && resp != nil {
		dumped, _ := httputil.DumpResponse(resp, false)
		log.Info(string(dumped))
	}

	if err != nil {
		return Response{
			Body:       err.Error(),
			StatusCode: 500,
			DhpCode:    0,
			Response:   resp,
		}

	}

	// Read Response Body
	responseBody, _ := ioutil.ReadAll(resp.Body)

	jsonParsed, err := gabs.ParseJSON([]byte(responseBody))
	if err == nil {
		dhpCode, ok := jsonParsed.Path("responseCode").Data().(string)
		if ok {
			if client.config.Debug {
				log.Info("Found responseCode: ", dhpCode)
			}
			intCode, _ := strconv.Atoi(dhpCode)
			return Response{
				Body:       string(responseBody),
				StatusCode: resp.StatusCode,
				DhpCode:    intCode,
				Response:   resp,
			}
		} else {
			log.Error("Response code not found")
		}

	}
	if client.config.Debug {
		log.Info("Returning RAW response")
	}
	return Response{
		Body:       string(responseBody),
		StatusCode: resp.StatusCode,
		DhpCode:    0,
		Response:   resp,
	}
}

func (client *ApiClient) sendSignedRequest(httpMethod, apiEndpoint, queryParams string, header *http.Header, body []byte) Response {
	now := time.Now().UTC()
	client.Sign(now, header, apiEndpoint, queryParams, httpMethod, body)
	uri := client.createUri(apiEndpoint, queryParams)

	return client.sendRestRequest(httpMethod, uri, header, body)
}

func (client *ApiClient) SendRestRequest(httpMethod, apiEndpoint, queryParams string, header *http.Header, body []byte) Response {
	uri := client.createUri(apiEndpoint, queryParams)
	return client.sendRestRequest(httpMethod, uri, header, body)
}

func (client *ApiClient) DHPApplicationName() string {
	return client.config.DhpApplicationName
}

// Sends a signed request to the configured service
// It returns a Response struct which contains the parsed response code
// from the service. A full copy of the body is also returned
func (client *ApiClient) SendSignedRequest(httpMethod, apiEndpoint, queryParams string, header *http.Header, body []byte) Response {

	return client.sendSignedRequest(httpMethod, apiEndpoint, queryParams, header, body)
}

func (client *ApiClient) Sign(now time.Time, header *http.Header, url, queryParams, httpMethod string, body []byte) {
	if header.Get("SignedDate") == "" {
		header.Set("SignedDate", now.Format(TIME_FORMAT))
	}
	authHeaderValue := client.apiSigner.BuildAuthorizationHeaderValue(httpMethod, queryParams, header, url, body)
	header.Set("Authorization", authHeaderValue)
}

func (client *ApiClient) createUri(apiEndpoint, queryParams string) *url.URL {
	log.Debug("BaseUrl: %s", client.apiBaseUrl)
	url, _ := url.Parse(client.apiBaseUrl)
	url.Parse(client.apiBaseUrl)
	url.Path = apiEndpoint
	url.RawQuery = queryParams
	return url
}
