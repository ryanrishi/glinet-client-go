glinet-client-go
===


## Usage

```go
import glinet "github.com/ryanrishi/glinet-client-go"
```

## Authentication Overview
```
$ curl -XPOST http://192.168.8.1/rpc -d '{"jsonrpc": "2.0", "id": 1, "method": "challenge", "params": {"username": "root"}}'
{"id":1,"jsonrpc":"2.0","result":{"salt":"1Aa2BbC3","alg":1,"nonce":"asdflkjasdflkj"}}
$ openssl passwd -1 -salt $salt $GLINET_PASSWORD | tee >hash
$ echo -n "root:$hash:$nonce" | md5sum | tee>login_hash
$ curl -XPOST http://192.168.8.1/rpc -d '{"jsonrpc": "2.0", "id": 1, "method": "login", "params": {"username": "root", "hash": "$login_hash"}}' | jq '.result'
{ "sid": ... }
```
