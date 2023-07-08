glinet-client-go
===
[![Tests](https://github.com/ryanrishi/glinet-client-go/actions/workflows/go.yml/badge.svg)](https://github.com/ryanrishi/glinet-client-go/actions/workflows/go.yml)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/ryanrishi/glinet-client-go)](https://github.com/ryanrishi/glinet-client-go/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/ryanrishi/glinet-client-go)](https://goreportcard.com/report/github.com/ryanrishi/glinet-client-go)


A Go client to access [GL.iNet](https://www.gl-inet.com/) routers. Based on [v4 firmware](https://dev.gl-inet.com/router-4.x-api/).


## Usage

### Installing
Use `go get` to retrieve the SDK to add it to your `GOPATH` workspace, or
project's Go module dependencies.

	go get github.com/ryanrishi/glinet-client-go

To update the SDK use `go get -u` to retrieve the latest version of the SDK.

	go get -u github.com/glinet-client-go


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
