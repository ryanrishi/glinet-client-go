package glinet

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
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

type loginRequest struct {
	Username string `json:"username"`
	Hash     string `json:"hash"`
}

type loginResponse struct {
	Sid string `json:"sid"`
}

func (s *DigestService) Challenge(username string) (*challengeResponse, error) {
	requestBody := challengeRequest{Username: username}
	var responseBody challengeResponse
	err := s.client.Call("challenge", &requestBody, &responseBody)

	if err != nil {
		return nil, err
	}

	return &responseBody, nil
}

func (s *DigestService) Login(username string, password []byte) (*loginResponse, error) {
	challenge, err := s.Challenge(username)
	if err != nil {
		return nil, err
	}

	hasher := md5.New()
	hasher.Write(password)
	hasher.Write([]byte(challenge.Salt))
	sum := hasher.Sum(nil)
	unixHash := make([]byte, len(sum)*2)
	hex.Encode(unixHash, hasher.Sum(nil))

	loginToBeHashed := md5.New().Sum([]byte(fmt.Sprintf("%s:%s:%s", username, unixHash, challenge.Nonce)))
	// loginHash example f62915b3ed48049e9d25dae6338b5dc9
	loginHash := make([]byte, len(loginToBeHashed)*2)
	hex.Encode(loginHash, loginToBeHashed)
	request := loginRequest{Username: username, Hash: string(loginHash)}
	var res loginResponse
	err = s.client.Call("login", &request, res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
