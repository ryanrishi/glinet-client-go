package glinet

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

const (
	Version          = "v0.0.1"
	defaultBaseUrl   = "http://192.168.8.1/rpc"
	defaultUserAgent = "glinet-client-go" + "/" + Version
)

type Client struct {
	client    *http.Client
	BaseURL   *url.URL
	UserAgent string
	common    service // Reuse a single struct instead of allocating one for each service on the heap.
	Digest    *DigestService
}

type jsonRcpRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type Response struct {
	Response *http.Response
	ID       int
	JSONRPC  string
	Result   interface{}
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	baseUrl, _ := url.Parse(defaultBaseUrl)

	c := &Client{client: httpClient, BaseURL: baseUrl, UserAgent: defaultUserAgent}
	c.Digest = (*DigestService)(&c.common)

	return c
}

func (c *Client) NewRequest(method string, params interface{}) (*http.Request, error) {
	var buf io.ReadWriter

	buf = &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	err := enc.Encode(&jsonRcpRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  method,
		Params:  params,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.BaseURL.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	return req, nil
}
