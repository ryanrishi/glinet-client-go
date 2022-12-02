package glinet

import (
	"bytes"
	"context"
	"errors"
	"github.com/gorilla/rpc/v2/json2"
	"net/http"
	"net/url"
)

const (
	Version          = "v0.0.1"
	defaultBaseUrl   = "http://192.168.8.1/rpc"
	defaultUserAgent = "glinet-client-go" + "/" + Version
)

type Client struct {
	BaseURL   *url.URL
	UserAgent string
	common    service // Reuse a single struct instead of allocating one for each service on the heap.
	Digest    *DigestService
}

var errNonNilContext = errors.New("ctx must not be nil")

type service struct {
	client  *Client
	context context.Context
}

func NewClient() *Client {
	baseUrl, _ := url.Parse(defaultBaseUrl)

	c := &Client{BaseURL: baseUrl, UserAgent: defaultUserAgent}
	c.common.client = c
	c.common.context = context.TODO()
	c.Digest = (*DigestService)(&c.common)

	return c
}

func (c *Client) Call(method string, params, result interface{}) error {
	buf, _ := json2.EncodeClientRequest(method, params)
	body := bytes.NewBuffer(buf)
	res, err := http.Post(c.BaseURL.String(), "application/json", body)
	if err != nil {
		return err
	}

	err = json2.DecodeClientResponse(res.Body, result)
	if err != nil {
		return err
	}

	return nil
}
