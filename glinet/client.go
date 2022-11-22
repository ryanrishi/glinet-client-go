package glinet

import (
	"context"
	"errors"
	"github.com/sourcegraph/jsonrpc2"
	"net"
)

type Client struct {
	host     net.Addr
	username string
	conn     jsonrpc2.Conn
}

type ChallengeParameters struct {
	username string
}

type ChallengeResponse struct {
	salt  string
	alg   int
	nonce string
}

type AlgoType int64

const (
	MD5 AlgoType = iota
	SHA256
	SHA512
	UKNOWN
)

func GetAlgoType(code int) (AlgoType, error) {
	switch code {
	case 1:
		return MD5, nil
	case 5:
		return SHA256, nil
	case 6:
		return SHA512, nil
	default:
		return UKNOWN, errors.New("unknown type")
	}
}

func (c *Client) Login(ctx context.Context) {
	params := ChallengeParameters{
		c.username,
	}

	var res ChallengeResponse
	err := c.conn.Call(ctx, "challenge", params, &res)
	if err != nil {
		return
	}

}
