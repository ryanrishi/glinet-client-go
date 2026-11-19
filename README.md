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

Each HTTP request, also the login, has a time limit of 60 s: the longest time that the web UI waits for the answer of a method of this client (`NetworkMode.SetMode`). To use a different limit or HTTP client, give `glinet.WithHTTPClient` to the constructor. Its `Timeout` then applies, and a zero `Timeout` gives no time limit. The deadline of the context of a call also applies.

```go
client, err := glinet.NewClientWithHost("192.168.8.1", "root", []byte(password),
	glinet.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}))
```

The client has the services `AccessControl`, `AdGuardHome`, `Bark`, `Cable`, `Clients`, `DDNS`, `Digest`, `DNS`, `Fan`, `Firewall`, `IGMP`, `IPv6`, `LAN`, `LED`, `LocalAccess`, `MultiWAN`, `Network`, `NetworkMode`, `ParentalControl`, `QoS`, `Repeater`, `SwitchButton`, `System`, `Tailscale`, `Tethering`, `Timer`, `VPNPolicy`, `Wifi`, `WireGuardClient`, `WireGuardServer` and `ZeroTier`.
The write methods of `VPNPolicy`, `WireGuardClient` and `WireGuardServer` were not tested on a router: their parameters come from the web UI code. The same applies to the DNS modes `manual`, `proxy` and `secure` of `DNS.SetConfig`; only mode `auto` was tested.
These parameters also come from the web UI code of firmware 4.8.3 and were not tested on a router: the DMZ `Priority` and `Firewall.OrderPortForwards`, the guest `WANIsolate` of `LAN.SetAddress`, the second WAN port (`glinet.SecondWANPort`) of `Cable` and `Network.CheckWANCable`, `SetConfig` of `AccessControl` and its allow list in mode `white`, `Chan6GOnlyPSC` of `Wifi.SetConfig`, `Firewall.SetWANAccess`, `LocalAccess.SetConfig`, the writes of `Bark`, `IGMP`, `MultiWAN`, `NetworkMode`, `ParentalControl`, `QoS`, `Repeater`, `SwitchButton` and `Tethering`, `Clients.SetInfo`, `Cable.SetPort`, `Network.SetAdvancedConfig`, `System.SetUSB3Disabled`, `Timer.SetScreenSchedule`, the "remove all" methods (`Firewall.RemoveAllPortForwards`, `Firewall.RemoveAllRules`, `LAN.RemoveAllStaticBinds`, `WireGuardServer.RemoveAllPeers`), and the empty values that `LAN.SetDHCP` (`lpr: []`) and the `WireGuardClient` and `WireGuardServer` writes send (see "Write formats from the web UI code" in `testdata/fixtures/axt1800_4.8.3/REPORT.md`). `Fan.SetTemperature`, `AccessControl.AddAllowedMAC` and `RemoveAllowedMAC` in mode `black`, and `Firewall.AddRule` with IPv6 addresses, CIDR blocks, the protocols `esp` and `icmpv6`, and an upper-case source MAC were tested on a router with firmware 4.8.3. `WireGuardClient` calls the `wg-client` module of firmware 4.7.0; firmware 4.9.0 replaces it with `vpn-client`. On firmware 4.8.3, `WireGuardClient.GetStatus`, `GetSetting`, `ListRoutes` and all reads of `VPNPolicy` return error -32601 (method not found); the tunnel status is in `vpn-client.get_status` (fixtures in `testdata/fixtures/axt1800_4.8.3`).
`NewClientUnauthenticated` returns a client that has no session. Only `Digest` works without a session.

`Call` calls a module method that has no typed method yet. A nil `params` sends `{}`. A nil `result` discards the response.

```go
var res struct {
	Device  string `json:"device"`
	Network string `json:"network"`
}
err := client.Call(ctx, "repeater", "get_status", nil, &res)
```

Router errors are `*glinet.RPCError`. This includes a result with a non-zero `err_code`, which some write methods return:

```go
var rpcErr *glinet.RPCError
if errors.As(err, &rpcErr) && rpcErr.Code == glinet.CodeMethodNotFound {
	// The firmware does not have this method.
}
```

An HTTP status that is not 200 is `*glinet.HTTPError`. It has the status code and the module and method of the call. For the login, `Module` is empty and `Method` is `challenge` or `login`. For example, the router returns HTTP status 500 when `firewall.remove_rule` gets an unknown ID:

```go
var httpErr *glinet.HTTPError
if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusInternalServerError {
	// The router returned HTTP status 500.
}
```

### Tests
Package `glinettest` starts a fake router with `httptest`. It implements `challenge`, `login` and `call`, and keeps the state of each module in memory.

```go
srv := glinettest.NewServer(t)
client, err := glinet.NewClientWithHost(srv.URL, srv.Username, []byte(srv.Password))

srv.ExpireSessions()                    // the next call gets -32000
srv.Handle("system", "get_info", method) // add or replace a method
srv.Handler("system", "get_info")        // the method that serves system.get_info now
srv.System.SetInfo(json.RawMessage(`{"model": "mt6000"}`)) // the result of a read-only method
```

The results of the read-only methods are those of the test router (`testdata/fixtures/axt1800_4.7.0`; 3 of its clients for `clients.get_list`) until a test sets them.

A method can return a `*glinettest.Error` to send a JSON-RPC error, or `glinettest.ErrInternalServerError` to send HTTP status 500.

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
