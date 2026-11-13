package glinet

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"github.com/GehirnInc/crypt/md5_crypt"
)

const algMD5Crypt = 1

type DigestService service

type challengeRequest struct {
	Username string `json:"username"`
}

type ChallengeResponse struct {
	Salt       string `json:"salt"`
	Algorithm  int    `json:"alg"`
	Nonce      string `json:"nonce"`
	HashMethod string `json:"hash_method,omitempty"`
}

type loginRequest struct {
	Username string `json:"username"`
	Hash     string `json:"hash"`
}

type LoginResponse struct {
	Username string `json:"username"`
	Sid      string `json:"sid"`
}

func (s *DigestService) Challenge(ctx context.Context, username string) (*ChallengeResponse, error) {
	var res ChallengeResponse
	if err := s.client.rpc(ctx, "challenge", challengeRequest{Username: username}, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Login gets a challenge and logs in with its nonce. The router accepts the nonce for 1000 ms.
func (s *DigestService) Login(ctx context.Context, username string, password []byte) (*LoginResponse, error) {
	challenge, err := s.Challenge(ctx, username)
	if err != nil {
		return nil, err
	}

	if challenge.Algorithm != algMD5Crypt || challenge.HashMethod != "" {
		return nil, fmt.Errorf("glinet: unsupported login algorithm (alg %d, hash_method %q): only alg %d (MD5-crypt) is supported",
			challenge.Algorithm, challenge.HashMethod, algMD5Crypt)
	}

	passwordHash, err := md5_crypt.New().Generate(password, []byte("$1$"+challenge.Salt))
	if err != nil {
		return nil, err
	}

	loginHash := md5.Sum([]byte(username + ":" + passwordHash + ":" + challenge.Nonce))

	var res LoginResponse
	err = s.client.rpc(ctx, "login", loginRequest{Username: username, Hash: hex.EncodeToString(loginHash[:])}, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
