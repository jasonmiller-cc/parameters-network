# parameters-network

REST API service for Linux network interface and routing management, built on `github.com/jasonmiller-cc/parameters-core`.

## Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/interfaces` | List all interfaces with stats |
| GET | `/api/v1/interfaces/{name}` | Get interface details |
| PUT | `/api/v1/interfaces/{name}` | Update interface (MTU, up/down) |
| GET | `/api/v1/interfaces/{name}/addresses` | List IP addresses on interface |
| POST | `/api/v1/interfaces/{name}/addresses` | Add IP address |
| DELETE | `/api/v1/interfaces/{name}/addresses/{address}` | Remove IP address |
| GET | `/api/v1/routes` | List routing table (`?table=&family=`) |
| POST | `/api/v1/routes` | Add route |
| DELETE | `/api/v1/routes` | Delete route |
| GET | `/api/v1/neighbors` | List ARP/NDP neighbors |
| GET | `/api/v1/vlans` | List VLAN interfaces |
| POST | `/api/v1/vlans` | Create VLAN interface |
| DELETE | `/api/v1/vlans/{name}` | Delete VLAN interface |
| GET | `/api/v1/stats` | Per-interface rx/tx byte counters |

## Building

```bash
# macOS / cross-compile
make build-linux

# Native Linux
make build
```

## Running

```bash
./bin/parameters-network -config /etc/parameters/network.yaml
```

Default port: **8086** (overridable via `PARAMS_NETWORK_SERVER_PORT`).

## Platform notes

Netlink operations require Linux. On non-Linux platforms the service compiles
via a build-tagged stub that returns a `not implemented` error on every call.
This keeps the compile/test loop fast on developer laptops.
