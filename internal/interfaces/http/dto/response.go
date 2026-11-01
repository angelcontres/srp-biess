package dto

// APIResponse es el sobre estándar para respuestas JSON de la API.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

func Success(data interface{}, message ...string) APIResponse {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	return APIResponse{
		Success: true,
		Data:    data,
		Message: msg,
	}
}

func Failure(errMsg string) APIResponse {
	return APIResponse{
		Success: false,
		Error:   errMsg,
	}
}
