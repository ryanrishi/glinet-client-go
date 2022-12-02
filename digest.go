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

//type challengeResponse struct {
//	ID      int               `json:"id"`
//	JSONRPC string            `json:"jsonrpc"`
//	Result  challengeResponse `json:"result"`
//}

//type RPCTransport struct {
//	// TODO parameterize this since all requests will have this structure (is my Java showing?)
//	ID      string            `json:"id"`
//	JSONRPC string            `json:"jsonrpc"`
//	Result  challengeResponse `json:"result"`
//}

func (s *DigestService) Challenge(ctx context.Context, username string) (*challengeResponse, error) {
	requestBody := challengeRequest{Username: username}
	var responseBody challengeResponse
	err := s.client.client.Call("challenge", &requestBody, &responseBody)

	if err != nil {
		return nil, err
	}

	return &responseBody, nil
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
