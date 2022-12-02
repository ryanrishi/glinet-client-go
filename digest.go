package glinet

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/GehirnInc/crypt/md5_crypt"
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
	Username string `json:"username"`
	Sid      string `json:"sid"`
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

	md5PasswordHash, err := md5_crypt.New().Generate(password, []byte(fmt.Sprintf("$1$%s", challenge.Salt)))
	if err != nil {
		return nil, err
	}

	// call challenge again because it sometimes times out
	challenge, err = s.Challenge(username)
	if err != nil {
		return nil, err
	}

	loginHash := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", username, md5PasswordHash, challenge.Nonce)))
	digest := make([]byte, hex.EncodedLen(len(loginHash)))
	hex.Encode(digest, loginHash[:])
	request := loginRequest{Username: username, Hash: string(digest)}
	var res loginResponse
	err = s.client.Call("login", &request, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
