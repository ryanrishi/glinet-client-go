package glinet

import (
	"bytes"
	"context"
	"github.com/gorilla/rpc/v2/json2"
	"log"
	"net/http"
	"net/url"
)

const (
	Version          = "v0.0.3"
	defaultBaseUrl   = "http://192.168.8.1/rpc"
	defaultUserAgent = "glinet-client-go" + "/" + Version
)

type Client struct {
	BaseURL   *url.URL
	UserAgent string
	common    service // Reuse a single struct instead of allocating one for each service on the heap.
	Sid       string

	// services
	AdGuard *AdGuardService
	Digest  *DigestService
	System  *SystemService
}

type NewClientParams struct {
	Username string
	Password []byte
}

type service struct {
	client  *Client
	context context.Context
}

func NewClient(params *NewClientParams) *Client {
	baseUrl, _ := url.Parse(defaultBaseUrl)

	c := &Client{BaseURL: baseUrl, UserAgent: defaultUserAgent}
	c.common.client = c
	c.common.context = context.TODO()

	// services
	c.AdGuard = (*AdGuardService)(&c.common)
	c.Digest = (*DigestService)(&c.common)
	c.System = (*SystemService)(&c.common)

	login, err := c.Digest.Login(params.Username, params.Password)
	if err != nil {
		log.Fatal("Error logging in: ", err)
	}

	c.Sid = login.Sid
	params.Password = nil

	return c
}

func NewClientUnauthenticated() *Client {
	baseUrl, _ := url.Parse(defaultBaseUrl)

	c := &Client{BaseURL: baseUrl, UserAgent: defaultUserAgent}
	c.common.client = c
	c.common.context = context.TODO()

	// services
	c.AdGuard = (*AdGuardService)(&c.common)
	c.Digest = (*DigestService)(&c.common)
	c.System = (*SystemService)(&c.common)

	return c
}

func (c *Client) CallWithStringSlice(method string, params []string, result interface{}) error {
	buf, _ := json2.EncodeClientRequest(method, params)
	err := c.call(buf, result)

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) CallWithInterface(method string, params, result interface{}) error {
	buf, _ := json2.EncodeClientRequest(method, params)
	err := c.call(buf, result)

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) CallWithInterfaceSlice(method string, params []interface{}, result interface{}) error {
	buf, _ := json2.EncodeClientRequest(method, params)
	err := c.call(buf, result)

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) call(buf []byte, result interface{}) error {
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
