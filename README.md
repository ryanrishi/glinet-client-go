glinet-client-go
===
[![Tests](https://github.com/ryanrishi/glinet-client-go/actions/workflows/go.yml/badge.svg)](https://github.com/ryanrishi/glinet-client-go/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/ryanrishi/glinet-client-go.svg)](https://pkg.go.dev/github.com/ryanrishi/glinet-client-go)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/ryanrishi/glinet-client-go)](https://github.com/ryanrishi/glinet-client-go/releases/latest)


A Go client to access [GL.iNet](https://www.gl-inet.com/) routers that run [v4 firmware](https://docs.gl-inet.com/router/en/4/).
API reference: [archived copy of the GL.iNet 4.x API documentation](https://web.archive.org/web/20240121142533/https://dev.gl-inet.com/router-4.x-api/).


## Usage

### Installing
Add the module to your project:

	go get github.com/ryanrishi/glinet-client-go@latest


### Client
`NewClient` and `NewClientWithHost` log in to the router and return an error if the login fails.
The client logs in again one time when a call returns `-32000` (access denied), for example after the session expired.

```go
ctx := context.Background()

client, err := glinet.NewClientWithHost("192.168.8.1", "root", []byte(password))
if err != nil {
	return err
}

tz, err := client.System.GetTimezoneConfig(ctx)
```

The client has the services `AdGuard`, `Digest`, `LAN` and `System`.
`NewClientUnauthenticated` returns a client that has no session. Only `Digest` works without a session.

`Call` calls a module method that has no typed method yet. A nil `params` sends `{}`. A nil `result` discards the response.

```go
var res struct {
	Interfaces []struct {
		Interface string `json:"interface"`
		IP        string `json:"ip"`
	} `json:"interfaces"`
}
err := client.Call(ctx, "lan", "get_config_list", nil, &res)
```

Router errors are `*glinet.RPCError`:

```go
var rpcErr *glinet.RPCError
if errors.As(err, &rpcErr) && rpcErr.Code == glinet.CodeMethodNotFound {
	// The firmware does not have this method.
}
```

### Tests
Package `glinettest` starts a fake router with `httptest`. It implements `challenge`, `login` and `call`, and keeps the state of each module in memory.

```go
srv := glinettest.NewServer(t)
client, err := glinet.NewClientWithHost(srv.URL, srv.Username, []byte(srv.Password))

srv.ExpireSessions()                    // the next call gets -32000
srv.Handle("system", "get_info", method) // add or replace a method
```

To run the read-only test against a real router, set `GLINET_USERNAME` and `GLINET_PASSWORD`, and `GLINET_HOST` if the router is not at `http://192.168.8.1` (in the environment or in `.env`) and run:

	GLINET_ROUTER_TEST=1 go test -run TestRouterReadOnly .

## Run examples
Most examples require authentication. Set `GLINET_USERNAME` and `GLINET_PASSWORD` environment variables.

If running through GoLand, I recommend using the [EnvFile plugin](https://plugins.jetbrains.com/plugin/7861-envfile).


## Authentication Overview
```sh
$ curl -XPOST http://192.168.8.1/rpc -d '{"jsonrpc": "2.0", "id": 1, "method": "challenge", "params": {"username": "root"}}'
{"id":1,"jsonrpc":"2.0","result":{"salt":"1Aa2BbC3","alg":1,"nonce":"asdflkjasdflkj"}}
$ openssl passwd -1 -salt $salt $GLINET_PASSWORD | tee >hash
$ echo -n "root:$hash:$nonce" | md5sum | tee>login_hash
$ curl -XPOST http://192.168.8.1/rpc -d '{"jsonrpc": "2.0", "id": 1, "method": "login", "params": {"username": "root", "hash": "$login_hash"}}' | jq '.result'
{ "sid": ... }
```

## Releasing
To release, change the `Version` constant in `glinet.go` and merge the change into `main`.
The `Release` workflow creates the tag and the GitHub release for a version that has no tag.
