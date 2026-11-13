package glinet

import "fmt"

const (
	CodeAccessDenied   = -32000
	CodeMethodNotFound = -32601
)

// RPCError is a JSON-RPC error object that the router returned.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("glinet: rpc error %d: %s", e.Code, e.Message)
}
