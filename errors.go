package glinet

import (
	"fmt"
	"net/http"
	"strings"
)

// JSON-RPC error codes that the router returns.
const (
	CodeAccessDenied   = -32000
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
)

// RPCError is an error that the router returned: a JSON-RPC error object, or the err_code and
// err_msg of a result, which some methods return, for example -1 "parameter missing".
// Err is the detail that some methods add to a JSON-RPC error, for example "invalid mac".
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     string `json:"err,omitempty"`
}

func (e *RPCError) Error() string {
	if e.Err != "" {
		return fmt.Sprintf("glinet: rpc error %d: %s: %s", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("glinet: rpc error %d: %s", e.Code, e.Message)
}

// HTTPError is an HTTP response with a status that is not 200. For example, the router returns
// HTTP status 500 when firewall.remove_rule gets an unknown ID.
// Module and Method are the module method of Call. For the challenge and the login, Module is
// empty and Method is "challenge" or "login".
// HTTPError does not keep the response body. On firmware 4.7.0 the body is an nginx HTML error page
// with no detail of the fault. The body on firmware 4.8.3 was not recorded.
type HTTPError struct {
	StatusCode int
	Module     string
	Method     string
}

func (e *HTTPError) Error() string {
	name := e.Method
	if e.Module != "" {
		name = e.Module + "." + e.Method
	}
	status := strings.TrimSpace(fmt.Sprintf("%d %s", e.StatusCode, http.StatusText(e.StatusCode)))
	return fmt.Sprintf("glinet: %s: unexpected HTTP status %s", name, status)
}
