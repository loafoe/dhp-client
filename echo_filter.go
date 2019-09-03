package client

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Jeffail/gabs/v2"
	"github.com/labstack/echo"
	"github.com/m4rw3r/uuid"
)

const (
	RESPONSE_CODE_GATEWAY_TIMEOUT       = 504
	RESPONSE_CODE_VALID_TOKEN           = 1152
	RESPONSE_CODE_TOKEN_EXPIRED         = 1008
	RESPONSE_CODE_TOKEN_INVALID         = 1009
	RESPONSE_CODE_INVALID_USER_ID       = 1004
	RESPONSE_CODE_ACCESS_TOKEN_REQUIRES = 1251
	RESPONSE_CODE_VALIDATION_ERRORS     = 1254
)

func DHPErrorResponse(errCode int, c echo.Context) {
	uuid, _ := uuid.V4()
	stringErrorCode := fmt.Sprintf("%d", errCode)
	response := &ErrorResponse{
		IncidentID:  uuid.String(),
		ErrorCode:   stringErrorCode,
		Description: StatusCodeToString(errCode),
	}
	c.JSON(400, response)
}

type tokenResponse struct {
	valid  bool
	reason int
}

func EchoAuthFilter(apiClients ...*ApiClient) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			GUID := c.Param("GUID")
			if GUID == "" {
				DHPErrorResponse(RESPONSE_CODE_INVALID_USER_ID, c)
				return echo.ErrUnauthorized
			}
			auth := c.Request().Header.Get(echo.HeaderAuthorization)
			if len(auth) == 0 {
				DHPErrorResponse(RESPONSE_CODE_ACCESS_TOKEN_REQUIRES, c)
				return echo.ErrUnauthorized
			}
			bearer := "bearer "
			l := len(bearer)
			if len(auth) > l+1 && strings.ToLower(auth[:l]) != bearer {
				DHPErrorResponse(RESPONSE_CODE_ACCESS_TOKEN_REQUIRES, c)
				return echo.ErrUnauthorized
			}
			t := auth[l:]
			bearerToken := string(t)
			ch := make(chan tokenResponse)
			timeout := time.After(30 * time.Second) // Wait max 10 seconds
			for _, client := range apiClients {
				go func(apiClient *ApiClient) {
					header := &http.Header{}
					method := "GET"
					apiEndpoint := "/authentication/users/" + GUID + "/tokenStatus"
					queryParams := "applicationName=" + apiClient.DHPApplicationName()
					var body []byte
					header.Add("AccessToken", bearerToken)
					response := apiClient.SendSignedRequest(method, apiEndpoint, queryParams, header, body)
					if response.StatusCode < 200 {
						// TODO
					}
					jsonParsed, err := gabs.ParseJSON([]byte(response.Body))
					if err != nil {
						DHPErrorResponse(RESPONSE_CODE_VALIDATION_ERRORS, c)
						ch <- tokenResponse{false, RESPONSE_CODE_VALIDATION_ERRORS}
						return
					}
					responseCode, ok := jsonParsed.Path("responseCode").Data().(string)
					if !ok {
						ch <- tokenResponse{false, RESPONSE_CODE_VALIDATION_ERRORS}
						return
					}
					intResponseCode, _ := strconv.ParseInt(responseCode, 10, 64)
					switch intResponseCode {
					case RESPONSE_CODE_VALID_TOKEN:
						break
					default:
						ch <- tokenResponse{false, int(intResponseCode)}
						return
					}
					ch <- tokenResponse{true, RESPONSE_CODE_VALID_TOKEN}
					return

				}(client)
			}
			results := 0
			for {
				select {
				case res := <-ch:
					if res.valid {
						return next(c)
					}
					results += 1
					if results >= len(apiClients) {
						DHPErrorResponse(res.reason, c)
						return echo.ErrUnauthorized
					}
					if res.reason == RESPONSE_CODE_TOKEN_EXPIRED ||
						res.reason == RESPONSE_CODE_TOKEN_INVALID {
						DHPErrorResponse(res.reason, c)
						return echo.ErrUnauthorized
					}
				case <-timeout:
					DHPErrorResponse(RESPONSE_CODE_GATEWAY_TIMEOUT, c)
					return echo.ErrUnauthorized

				}
			}
		}
	}
}
