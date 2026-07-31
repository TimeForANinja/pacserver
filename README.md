# PAC-Server

PAC-Server serves Proxy Auto Config files based on the client's source IP. It maps IP networks to PAC templates, supports template variables inside PAC files, and exposes request and event logging for operations work.

## What It Does

- Serves different PAC files by client network
- Resolves requests through a zone-to-template lookup tree
- Renders PAC templates with `Filename` and `Contact`
- Exposes Prometheus metrics when enabled
- Supports reload without restarting the server
- Logs every request to the access log and errors or panics to the event log

## Runtime Modes

The binary supports three flags:

```bash
pacserver --serve
pacserver --test
pacserver --reload
```

`--serve` starts the prod and admin listeners. `--test` loads the config, zones, and PAC files, then exits after validation. `--reload` sends a reload request to the admin listener using the configured admin secret.

## Request Routes

- Paths listed in `routes.csv` serve their configured PAC (including `/wpad.dat`)
- `/` resolves the best PAC for the request source IP on the prod listener
- `/:ip` resolves a specific IPv4 address as `/32`
- `/:ip/:cidr` resolves a specific IPv4 network prefix

Append `?debug=1` to any PAC route to return the matched request, lookup path, and PAC body in a debug response.

## Admin Routes

- `GET /admin` renders the admin UI or a login prompt
- `POST /admin/login` stores the admin secret in a cookie
- `POST /admin/reload` triggers a live reload

The admin endpoints live on the admin listener and require the configured `adminSecret`.
Prometheus metrics are exposed on the admin listener and observe only the prod listener's request traffic.

## Configuration

The application expects `config.yml` in the current working directory.

| Field               | Type   | Default                  | Description                                             |
|---------------------|--------|--------------------------|---------------------------------------------------------|
| `ipMapFile`         | string | `data/zones.csv`         | CSV file mapping IP networks to PAC files               |
| `routeMapFile`      | string | `data/routes.csv`        | CSV file mapping request paths to PAC files              |
| `pacRoot`           | string | `data/pacs`              | Directory containing PAC templates                      |
| `defaultPACFile`    | string | `${pacRoot}/default.pac` | PAC served when no zone matches                         |
| `contactInfo`       | string | `Your Help Desk`         | Contact text injected into PAC templates                |
| `accessLogFile`     | string | `access.log`             | JSON-lines request access log file                      |
| `eventLogFile`      | string | `event.log`              | Application event log file                              |
| `port`              | uint16 | `8080`                   | Prod listener port                                      |
| `adminPort`         | uint16 | `8082`                   | Admin listener port, including `/metrics`               |
| `adminACLs`         | string | `127.0.0.1/32`           | Comma-separated source networks allowed on admin port   |
| `adminSecret`       | string | empty                    | Shared secret for admin and reload endpoints            |
| `prometheusEnabled` | bool   | `false`                  | Enable Prometheus metrics                               |
| `prometheusPath`    | string | `/metrics`               | Metrics endpoint path                                   |
| `ignoreMinors`      | bool   | `false`                  | Continue startup when only minor load issues were found |
| `loglevel`          | string | `INFO`                   | Event-log level: `DEBUG`, `INFO`, `WARN`, or `ERROR`    |

## Zones CSV

The zones file is a CSV with no header. Blank lines and lines starting with `//` or `#` are ignored.

| Column | Type | Description |
| --- | --- | --- |
| `0` | ip | Network address |
| `1` | int | CIDR prefix length |
| `2` | file | PAC file path relative to `pacRoot` |
| `3` | text | Optional comment |

## Routes CSV

The routes file is a headerless CSV using `path, PAC file, optional comment`. Paths may be
written with or without surrounding slashes. For example:

```csv
wpad.dat, wpad.dat, WPAD discovery route
team/proxy.pac, other/team.pac, Team-specific route
```

Example:

```csv
192.168.1.0,24,default.pac
10.0.0.0,8,internal.pac
172.16.0.0,16,vpn.pac
```

## PAC Templates

PAC templates are rendered with the following variables:

| Variable | Meaning |
| --- | --- |
| `Filename` | Relative PAC filename |
| `Contact` | Contact text from `config.yml` |

Example:

```js
// This is the {{ .Filename }} PAC file
// For changes please reach out to {{ .Contact }}

function FindProxyForURL(url, host) {
    return "DIRECT";
}
```

## Logging And Diagnostics

- Every request is written to the access log as a single JSON line with the request method, path, status, latency, settled IP, raw `X-Forwarded-For`, and served PAC filename.
- Application logs go to stdout and the event log when serving.
- Unexpected panics are recovered and logged with a stack trace.
- Request handler errors and admin reload client failures are logged with contextual stack traces.

## Flow

The current startup flow is:

1. Load and validate `config.yml`
2. Load zones and PAC templates into the lookup cache
3. Start the prod and admin listeners for `--serve`
4. Trigger a live reload for `--reload`
5. Exit after validation for `--test`

The flow diagram in `docs/flow.drawio` tracks the same control flow.

## Building

```bash
go build -o pacserver ./cmd/pacserver.go
```

## Testing

```bash
go test ./...
```

## Project Layout

```text
pacserver/
├── cmd/           # CLI entrypoint
├── internal/      # Server bootstrap, config, storage, and request routing
├── pkg/IP/        # IPv4 and CIDR helpers
├── pkg/IPLUT/     # Generic IP lookup tree
├── pkg/admin/     # Admin UI and reload endpoints
├── pkg/utils/     # Shared helpers
├── docs/          # Flow diagram and docs assets
└── demo_files/    # Example zones and PAC files
```
