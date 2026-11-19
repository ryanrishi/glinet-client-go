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

## Naming

Follow these rules for each new exported name, so that the API stays consistent.

- Services: a service has a readable name of its router module (`LocalAccess` for `local-access`, `MultiWAN` for `kmwan`). The `Client` field and the `glinettest.Server` field have the same name.
- Read methods: `Get<X>` returns one object as `*T`. `List<X>` returns a slice of items. `Find<X>` returns nil and no error when the item does not exist; a `Get` method never does that. A method name does not repeat the service name: use `GetConfig`, `SetConfig` and `GetStatus` for the main configuration and status of a module.
- Write methods: `Add<X>` adds an item and returns its ID, `Set<X>` replaces, `Remove<X>` and `RemoveAll<X>` remove. Do not use Delete or Update. A `Set` method takes its argument as a value, not as a pointer. A write of one value can have a typed setter, for example `SetTemperature`.
- Result types: name a read result for the thing (`<Module>Config`, `<Module>Status`, `<Module>Info`), never `Get...Response`. When the write sends the same fields as the read, one type is used for both.
- Write types: when the params of a `SetConfig` method differ from the read type `<X>Config`, the params type is `Set<X>ConfigRequest`. A type with the values that more than one write sends, for example an add and a set, is `<X>Settings`.
- Type prefixes: the types of a service file start with the service name (`MultiWAN...`, `WireGuardServer...`). A short name that is not clear in the package, such as `Rule`, gets the prefix (`FirewallRule`).
- Fields: a field name is the JSON key, with each word capitalized and the Go initialisms in capitals (`DestIP`, `PeerID`, `SSHPort`, `MACAddr` for `macaddr`). Use a clearer name only when the key is not clear, and give the key in the doc comment (`Running` for `status`, `Interval` for `val`). The ID of an item is `ID`, also when the key is `group_id` or `peer_id`.
- Bool fields: a router value that means true or false is a `bool`, also when the router sends 0 and 1; then the type has a `MarshalJSON` and an `UnmarshalJSON` that use the unexported type `flag`. A value that has or can have a third state stays a number, and its doc comment says why. The bool that switches a feature or an item on is `Enabled`, also when the key is `enable` (`IPv6Enabled` for `ipv6_enable`). Keep a key that starts with `enable_` as `Enable<X>` (`EnableHTTPS`).
- Constants: a constant starts with its service or type name and then names the field that it is a value of (`MultiWANTrackProtoIPv4` for `TrackProto`, `MultiWANStatusOnline` for the status). `FirstWANPort` and `SecondWANPort` are the exception.
- `glinettest`: a state type and an accessor have the name of the root type or method (`Firewall.Zones` for `FirewallService.GetZones`, `MultiWANSensitivity.Interval`). An accessor pair is `X()` and `SetX()`.

## Releases

A merge to `main` that changes `Version` in `glinet.go` makes a tag and a release. Do not change `Version` in a pull request unless a maintainer asks for it.
