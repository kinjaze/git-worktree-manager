package jsonapi

type Error struct {
	Code             string `json:"code"`
	Message          string `json:"message"`
	LocalizedMessage string `json:"localizedMessage,omitempty"`
}

type Response struct {
	OK     bool   `json:"ok"`
	Status string `json:"status"`
	Error  *Error `json:"error,omitempty"`
	Data   any    `json:"data,omitempty"`
}

func Success(status string, data any) Response {
	return Response{OK: true, Status: status, Data: data}
}

func Failure(status string, code string, message string, localizedMessage string, data any) Response {
	return Response{
		OK:     false,
		Status: status,
		Error: &Error{
			Code:             code,
			Message:          message,
			LocalizedMessage: localizedMessage,
		},
		Data: data,
	}
}
