package glinet

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"

	"github.com/GehirnInc/crypt/md5_crypt"
)

const algMD5Crypt = 1

// loginHashes has the hash of the login for each hash-method of a challenge. A challenge
// without hash-method uses MD5.
var loginHashes = map[string]func() hash.Hash{
	"":       md5.New,
	"md5":    md5.New,
	"sha256": sha256.New,
}

// DigestService calls the challenge and login methods.
type DigestService service

type challengeRequest struct {
	Username string `json:"username"`
}

// Challenge is the result of challenge.
type Challenge struct {
	Salt       string `json:"salt"`
	Algorithm  int    `json:"alg"`
	Nonce      string `json:"nonce"`
	HashMethod string `json:"hash-method,omitempty"`
}

type loginRequest struct {
	Username string `json:"username"`
	Hash     string `json:"hash"`
}

// LoginResponse is the result of login. SID is the session ID.
type LoginResponse struct {
	Username string `json:"username"`
	SID      string `json:"sid"`
}

// Challenge returns the salt, algorithm and nonce for a login of username.
func (s *DigestService) Challenge(ctx context.Context, username string) (*Challenge, error) {
	var res Challenge
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

	loginHash, ok := loginHashes[challenge.HashMethod]
	if challenge.Algorithm != algMD5Crypt || !ok {
		return nil, fmt.Errorf("glinet: unsupported login algorithm (alg %d, hash-method %q): only alg %d (MD5-crypt) with hash-method md5 or sha256 is supported",
			challenge.Algorithm, challenge.HashMethod, algMD5Crypt)
	}

	passwordHash, err := md5_crypt.New().Generate(password, []byte("$1$"+challenge.Salt))
	if err != nil {
		return nil, err
	}

	h := loginHash()
	h.Write([]byte(username + ":" + passwordHash + ":" + challenge.Nonce))

	var res LoginResponse
	err = s.client.rpc(ctx, "login", loginRequest{Username: username, Hash: hex.EncodeToString(h.Sum(nil))}, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
