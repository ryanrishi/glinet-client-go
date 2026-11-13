package glinet

import "fmt"

// JSON-RPC error codes that the router returns.
const (
	CodeAccessDenied   = -32000
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
)

// RPCError is a JSON-RPC error object that the router returned.
// Err is the detail that some methods add, for example "invalid mac".
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
