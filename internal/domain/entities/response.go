package entities

type StandardResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type ErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Code    string `json:"code" example:"INVALID_REQUEST"`
	Error   string `json:"error" example:"error message description"`
	Details string `json:"details,omitempty"`
}
