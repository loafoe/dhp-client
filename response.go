package client

import (
	"net/http"
)

// Response is returned as the result of DHP request done with SendRestRequest() or SendSignedRequest()
// It contains the parsed response code and the full body. A list of error codes is available
type Response struct {
	StatusCode int            // The HTTP status code
	DhpCode    int            // The DHP response code
	Body       string         // The full body of the response
	Response   *http.Response // Useful when you need a http.Response representation
	Errors     []error        // Slice of identified errors in the response
}
