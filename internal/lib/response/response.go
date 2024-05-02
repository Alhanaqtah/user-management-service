package response

const (
	StatusOK  = "OK"
	StatusErr = "Error"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type Tokens struct {
	AccessToken  string `json:"accessToken,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
}

func Ok() Response {
	return Response{
		Status: StatusOK,
	}
}

func Err(errMsg string) Response {
	return Response{
		Status: StatusErr,
		Error:  errMsg,
	}
}
