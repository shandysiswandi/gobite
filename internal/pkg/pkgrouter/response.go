package pkgrouter

type errorResponse struct {
	Message string            `json:"message"`
	Error   map[string]string `json:"error,omitempty"`
}

type successReponse struct {
	Message string            `json:"message"`
	Data    any               `json:"data"`
	Meta    map[string]string `json:"meta,omitempty"`
}
