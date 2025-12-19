package handlers

type Response struct {
	Message     string `json:"message"`
	Data        any    `json:"data"`
	Error       any    `json:"error"`
	Info        any    `json:"info"`
	DevResponse any    `json:"dev_response"`
}

func NewSuccessResponse(message string, data any, infos ...any) Response {
	var info any
	if len(infos) > 0 {
		info = infos[0]
	}
	return Response{
		Message:     message,
		Data:        data,
		Error:       nil,
		Info:        info,
		DevResponse: nil,
	}
}

func NewErrorResponse(message string, err any, devResponses ...any) Response {
	var devResponse any
	if len(devResponses) > 0 {
		devResponse = devResponses[0]
	}
	return Response{
		Message:     message,
		Data:        nil,
		Error:       err,
		Info:        nil,
		DevResponse: devResponse,
	}
}
