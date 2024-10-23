package http

type Result struct {
	Code int         `json:"code,omitempty"`
	Msg  string      `json:"msg,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

func (r *Result) WithMsg(message string) *Result {
	r.Msg = message
	return r
}

func (r *Result) WithData(data interface{}) *Result {
	r.Data = data
	return r
}

func newResult(code int, msg string) *Result {
	return &Result{
		Code: code,
		Msg:  msg,
		Data: nil,
	}
}

func OK(msg string) *Result {
	return newResult(200, msg)
}

func Err(msg string) *Result {
	return newResult(500, msg)
}

func UnAuth(msg string) *Result {
	return newResult(401, msg)
}
