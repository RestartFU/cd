package rmux

type Response struct {
	Message    string
	StatusCode int
}

func NewResponse(message string, statusCode int) *Response {
	return &Response{
		Message:    message,
		StatusCode: statusCode,
	}
}
