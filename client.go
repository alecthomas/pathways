package pathways

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client request arguments.
type Args map[string]string

type ClientError struct {
	status int
	err    string
}

func (c *ClientError) StatusCode() int {
	return c.status
}

func (c *ClientError) Error() string {
	return c.err
}

// A HTTP client that uses named routes on a service to reconstruct and send requests.
type Client struct {
	service  *Service
	encoding string
	Client   *http.Client
}

// Create a new service client.
func NewClient(service *Service, encoding string) *Client {
	return &Client{
		service:  service,
		encoding: encoding,
		Client:   &http.Client{},
	}
}

// Call an API endpoint.
func (c *Client) Call(name string, args Args, request interface{}, response interface{}) (*http.Response, error) {
	// Encode the body
	bodyw := &bytes.Buffer{}
	err := Serializers.Encode(c.encoding, bodyw, request)
	if err != nil {
		return nil, err
	}
	body := bodyw.Bytes()

	req, err := c.MakeRequest(name, args, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return resp, &ClientError{
			status: resp.StatusCode,
			err:    fmt.Sprintf("HTTP error (%d): %s", resp.StatusCode, resp.Status),
		}
	}

	// Decode response
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, c.encoding) {
		return nil, fmt.Errorf("expected %s response from %s, got %s", c.encoding, req.URL, ct)
	}
	return resp, Serializers.Decode(c.encoding, resp.Body, response)
}

func (c *Client) MakeRequest(name string, args Args, body []byte) (*http.Request, error) {
	// Send the request
	route := c.service.Find(name)
	url := route.Reverse(args)
	method := route.Method()

	var content io.Reader
	if body != nil {
		content = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, content)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", c.encoding)
	req.Header.Set("Accept", c.encoding)
	return req, nil
}
