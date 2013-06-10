package pathways

import (
	"net/http"
)

// Request context.
type Context struct {
	Request  *http.Request
	Response http.ResponseWriter

	// Variables extracted from the URL path.
	PathVars map[string]string
	// User-defined variables.
	Vars map[string]interface{}
}

func (c *Context) InferContentType(defaultContentType string) string {
	ct := c.Request.Header.Get("Accept")
	if ct == "*/*" || ct == "" {
		if defaultContentType != "" {
			ct = defaultContentType
		} else {
			ct = c.Request.Header.Get("Content-Type")
		}
	}
	return ct
}

// Render a template.
func (c *Context) ApiError(code int, error string) Response {
	return c.ApiResponse(code, &ApiError{
		Status: code,
		Error:  error,
	})
}

func (c *Context) ApiResponse(code int, response interface{}) Response {
	return func(w http.ResponseWriter) {
		contentType := c.InferContentType("application/json")
		Serializers.EncodeResponse(w, code, contentType, response)
	}
}
