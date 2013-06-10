package pathways

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Args map[string]string

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

func (c *Client) Call(name string, args map[string]string, request interface{}, response interface{}) error {
	// Encode the body
	bodyw := &bytes.Buffer{}
	err := Serializers.Encode(c.encoding, bodyw, request)
	if err != nil {
		return err
	}
	body := bodyw.Bytes()

	req, err := c.MakeRequest(name, args, body)
	if err != nil {
		return err
	}

	resp, err := c.Client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}
	// Decode response
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, c.encoding) {
		return fmt.Errorf("expected %s response from %s, got %s", c.encoding, req.URL, ct)
	}
	return Serializers.Decode(c.encoding, resp.Body, response)
}

func (c *Client) MakeRequest(name string, args map[string]string, body []byte) (*http.Request, error) {
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
	req.Close = true
	return req, nil
}

func (c *Client) Do(name string, args Args, body []byte) (*http.Response, error) {
	req, err := c.MakeRequest(name, args, body)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}
