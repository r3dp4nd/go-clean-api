package server

type HomeResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
