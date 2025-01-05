package glinet

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/rpc/v2/json2"
)

const (
	Version          = "v0.0.4"
	defaultBaseUrl   = "http://192.168.8.1"
	defaultUserAgent = "glinet-client-go" + "/" + Version
)

type ClientInterface interface {
	GetSid() string
	CallWithStringSlice(method string, params []string, result interface{}) error
	CallWithInterface(method string, params, result interface{}) error
	CallWithInterfaceSlice(method string, params []interface{}, result interface{}) error
}

type Client struct {
	BaseURL   *url.URL
	UserAgent string
	common    service // Reuse a single struct instead of allocating one for each service on the heap.
	sid       string

	// services
	AdGuard *AdGuardService
	Digest  *DigestService
	System  *SystemService
}

type service struct {
	client  ClientInterface
	context context.Context
}

func NewClient(username string, password []byte) *Client {
	return NewClientWithHost(defaultBaseUrl, username, password)
}

func NewClientWithHost(host string, username string, password []byte) *Client {
	if !strings.HasPrefix(host, "http") {
		host = "http://" + host
	}

	baseUrl, err := url.Parse(host + "/rpc")

	if err != nil {
		log.Fatal("Error parsing host", err)
	}

	c := &Client{BaseURL: baseUrl, UserAgent: defaultUserAgent}
	c.common.client = c
	c.common.context = context.TODO()

	// services
	c.AdGuard = (*AdGuardService)(&c.common)
	c.Digest = (*DigestService)(&c.common)
	c.System = (*SystemService)(&c.common)

	login, err := c.Digest.Login(username, password)
	if err != nil {
		log.Fatal("Error logging in: ", err)
	}

	c.sid = login.Sid

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

func (c *Client) GetSid() string {
	return c.sid
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
