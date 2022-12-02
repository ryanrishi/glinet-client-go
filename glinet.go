package glinet

import (
	"errors"
	"log"
	"net/rpc"
	"net/url"
)

const (
	Version          = "v0.0.1"
	defaultBaseUrl   = "http://192.168.8.1/rpc"
	defaultUserAgent = "glinet-client-go" + "/" + Version
)

type Client struct {
	client    *rpc.Client
	BaseURL   *url.URL
	UserAgent string
	common    service // Reuse a single struct instead of allocating one for each service on the heap.
	Digest    *DigestService
}

//type jsonRcpRequest struct {
//	JSONRPC string      `json:"jsonrpc"`
//	ID      int         `json:"id"`
//	Method  string      `json:"method"`
//	Params  interface{} `json:"params"`
//}
//
//type Response struct {
//	*http.Response
//	ID      int
//	JSONRPC string
//	Result  interface{}
//}

var errNonNilContext = errors.New("ctx must not be nil")

//func newResponse(r *http.Response) *Response {
//	response := &Response{Response: r}
//	response.populateJsonRpcFields()
//	return response
//}

//func (r *Response) populateJsonRpcFields() {
//	// TODO can I read body multiple times?
//}

//func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*Response, error) {
//	if ctx == nil {
//		return nil, errNonNilContext
//	}
//
//	req.WithContext(ctx)
//
//	resp, err := c.client.Do(req)
//	if err != nil {
//		// If we got an error, and the context has been canceled,
//		// the context's error is probably more useful.
//		select {
//		case <-ctx.Done():
//			return nil, ctx.Err()
//		default:
//		}
//
//		// If the error type is *url.Error, sanitize its URL before returning.
//		if e, ok := err.(*url.Error); ok {
//			if url, err := url.Parse(e.URL); err == nil {
//				e.URL = url.String()
//				return nil, e
//			}
//		}
//
//		return nil, err
//	}
//
//	response := newResponse(resp)
//	// TODO more thorough error handling
//
//	switch v := v.(type) {
//	case nil:
//	case io.Writer:
//		_, err = io.Copy(v, response.Body)
//	default:
//		decErr := json.NewDecoder(resp.Body).Decode(response)
//		if decErr == io.EOF {
//			decErr = nil // ignore EOF errors caused by empty response body
//		}
//		if decErr != nil {
//			err = decErr
//		}
//	}
//
//	return response, err
//}

func NewClientWithAddress(addr string) *Client {
	baseUrl, _ := url.Parse(defaultBaseUrl)

	//conn, err := jsonrpc.Dial("tcp", addr)
	//if err != nil {
	//	log.Fatal(err)
	//}

	rpcClient, err := rpc.DialHTTPPath("tcp", addr, "/rpc")
	if err != nil {
		log.Fatal(err)
	}

	c := &Client{client: rpcClient, BaseURL: baseUrl, UserAgent: defaultUserAgent}
	c.common.client = c
	c.Digest = (*DigestService)(&c.common)

	return c
}

func NewClient() *Client {
	return NewClientWithAddress("192.168.8.1:80")
}

//func (c *Client) NewRequest(method string, params interface{}) (*http.Request, error) {
//	var buf io.ReadWriter
//
//	buf = &bytes.Buffer{}
//	enc := json.NewEncoder(buf)
//	enc.SetEscapeHTML(false)
//	err := enc.Encode(&jsonRcpRequest{
//		JSONRPC: "2.0",
//		ID:      1,
//		Method:  method,
//		Params:  params,
//	})
//	if err != nil {
//		return nil, err
//	}
//
//	req, err := http.NewRequest("POST", c.BaseURL.String(), buf)
//	if err != nil {
//		return nil, err
//	}
//
//	req.Header.Set("Content-Type", "application/json")
//	req.Header.Set("Accept", "application/json")
//
//	if c.UserAgent != "" {
//		req.Header.Set("User-Agent", c.UserAgent)
//	}
//
//	return req, nil
//}
