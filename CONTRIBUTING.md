# Contributing

Thank you for your help. This page gives the rules that are not obvious from the code.

## Use only made-up data

Code, tests, fixtures, documentation, commit messages and pull requests must contain only made-up values. Do not use values from your own router or network, also when they are not secret.

- IP addresses: use `192.168.8.x` (the router default) or the documentation ranges `192.0.2.0/24`, `198.51.100.0/24`, `203.0.113.0/24` and `2001:db8::/32`.
- MAC addresses: use made-up addresses such as `02:00:00:00:00:01`.
- Names: use names such as `example`. Do not use a real SSID, hostname, VPN peer name or DDNS name.
- Never include a password, a Wi-Fi key, a WireGuard key, a serial number or a session ID.

`cmd/probe` records the read results of a router as fixtures and replaces configured values. Read each fixture before you commit it: the redaction does not know each value that you configured.

When you post a router result in an issue, replace these values first. The structure, the key names and the value types are sufficient.

## Run the checks

You need Go (version in `.tool-versions`).

```bash
gofmt -l .
go vet ./...
go test ./...
```

The tests use the fake router of the package `glinettest`. They do not need a router.

## Rules for a new method

- A typed method has typed request and result structs, a doc comment, a handler in `glinettest`, and tests that check the exact parameters of each write.
- A write must be able to send each value that the router web UI can send, also an empty value.
- The doc comment states where the format comes from: a test on a router, a fixture, or the web UI code. It states what is not known. Do not guess a parameter.
- The `REPORT.md` files in `testdata/fixtures/` record the verified behaviour of a router and the formats that come from the web UI code. Add your findings there.

## Releases

A merge to `main` that changes `Version` in `glinet.go` makes a tag and a release. Do not change `Version` in a pull request unless a maintainer asks for it.
