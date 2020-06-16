package model

type Response struct {
	Data       interface{} `json:"data"`
	TotalCount int64       `json:"total_count,omitempty"`
	Detail     string      `json:"detail,omitempty"`
	Errors     interface{} `json:"errors,omitempty"`
}

func NewResponse() *Response {
	return &Response{}
}
