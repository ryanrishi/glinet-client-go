package glinet

import (
	"context"
)

type DigestService service

//type ChallengeRequest struct {
//	Username   string
//	Method     string
//	URI        string
//	Body       string
//	HTTPClient *http.Client
//}

type challengeRequest struct {
	Username string `json:"username"`
}

type challengeResponse struct {
	Salt      string `json:"salt"`
	Algorithm string `json:"alg"`
	Nonce     string `json:"nonce"`
}

type RPCTransport struct {
	// TODO parameterize this since all requests will have this structure (is my Java showing?)
	ID      string            `json:"id"`
	JSONRPC string            `json:"jsonrpc"`
	Result  challengeResponse `json:"result"`
}

func (s *DigestService) Challenge(ctx context.Context, username string) (*challengeResponse, *Response, error) {
	req, err := s.client.NewRequest("challenge", &challengeRequest{
		Username: username,
	})

	if err != nil {
		return nil, nil, err
	}

	c := new(challengeResponse)
	resp, err := s.client.Do(ctx, req, c)
	if err != nil {
		return nil, resp, err
	}

	return c, resp, nil
}

//func (cr *ChallengeRequest) Execute() (resp *http.Response, err error) {
//	var req *http.Request
//	if req, err = http.NewRequest(cr.Method, cr.URI, bytes.NewReader([]byte(cr.Body))); err != nil {
//		return nil, err
//	}
//
//	if resp, err = cr.HTTPClient.Do(req); err != nil {
//		return nil, err
//	}
//
//	return resp, nil
//}
