package glinet

import (
	"context"
)

type DigestService service

type challengeRequest struct {
	Username string `json:"username"`
}

type challengeResponse struct {
	Salt      string `json:"salt"`
	Algorithm int    `json:"alg"`
	Nonce     string `json:"nonce"`
}

func (s *DigestService) Challenge(ctx context.Context, username string) (*challengeResponse, error) {
	requestBody := challengeRequest{Username: username}
	var responseBody challengeResponse
	err := s.client.Call("challenge", &requestBody, &responseBody)

	if err != nil {
		return nil, err
	}

	return &responseBody, nil
}
