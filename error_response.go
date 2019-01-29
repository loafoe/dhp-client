package client

type ErrorResponse struct {
	IncidentID  string `json:"incidentID"`
	ErrorCode   string `json:"errorCode"`
	Description string `json:"description"`
}
