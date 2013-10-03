package pathways

import (
	"html/template"
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
	// Template, if any.
	Template *template.Template
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
func (c *Context) APIError(code int, error string) *Response {
	return c.APIResponse(code, &APIError{
		Status: code,
		Error:  error,
	})
}

func (c *Context) APIResponse(code int, response interface{}) *Response {
	return ResponseFromContext(c, func(w http.ResponseWriter) {
		contentType := c.InferContentType("application/json")
		Serializers.EncodeResponse(w, code, contentType, response)
	})
}

func (c *Context) Error(code int, error string) *Response {
	return ResponseFromContext(c, func(w http.ResponseWriter) {
		http.Error(w, error, code)
	})
}

// Render the template associated with this request.
func (c *Context) Render(data interface{}) *Response {
	return ResponseFromContext(c, func(w http.ResponseWriter) {
		err := c.Template.ExecuteTemplate(w, c.Template.Name(), data)
		if err != nil {
			c.Error(http.StatusInternalServerError, "Internal Server Error")
		}
	})
}
