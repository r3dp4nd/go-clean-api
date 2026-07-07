package server

type HomeResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

type PingResponse struct {
	Message   string `json:"message"`
	Version   string `json:"version"`
	RequestID string `json:"request_id"`
}

type ErrorResponse struct {
	Error APIError `json:"error"`
}

type APIError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}
