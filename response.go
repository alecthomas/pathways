package pathways

import (
	"net/http"
)

type ResponseWriter func(http.ResponseWriter)

type Response struct {
	Response http.ResponseWriter
	Request  *http.Request
	writer   ResponseWriter
}

func ResponseFromContext(cx *Context, writer ResponseWriter) *Response {
	return &Response{
		Request:  cx.Request,
		Response: cx.Response,
		writer:   writer,
	}
}

func (r *Response) ContentType(ct string) *Response {
	return r.Header("Content-Type", ct)
}

func (r *Response) Header(key, value string) *Response {
	r.Response.Header().Add(key, value)
	return r
}

func (r *Response) Write() {
}
