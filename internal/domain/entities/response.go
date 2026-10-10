package entities

type StandardResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type ErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Error   string `json:"error" example:"error message description"`
	Details string `json:"details,omitempty"`
}

