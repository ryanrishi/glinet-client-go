package glinet

import "fmt"

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
