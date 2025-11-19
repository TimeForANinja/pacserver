# PAC-Server
This is a custom server made to serve Proxy Auto Config (short PAC) Files based on Source IP.

## Setup
The setup of this App is pretty straight forward.
After downloading the executable, you only require the following three parameters:

### Config
Configuration is now provided via environment variables. All variables are prefixed with `APP_`.

Supported variables and defaults:

| Variable               | Type  | Default                 | Description                                                     |
|------------------------|-------|-------------------------|-----------------------------------------------------------------|
| `APP_IP_MAP_FILE`      | str   | `demo_files/zones.csv`  | Path to the Zones `.csv` file                                   |
| `APP_PAC_ROOT`         | str   | `demo_files/pacs`       | Path to the directory containing the PAC files                   |
| `APP_CONTACT_INFO`     | str   | ``                      | Contact info for use inside PAC templates                        |
| `APP_ACCESS_LOG_FILE`  | str   | `./access.log`          | Path to the access log file                                      |
| `APP_EVENT_LOG_FILE`   | str   | `./events.log`          | Path to the event log file                                       |
| `APP_DO_AUTO_REFRESH`  | bool  | `false`                 | Automatically refresh PAC and Zones on an interval               |
| `APP_MAX_CACHE_AGE`    | int   | `0`                     | Interval in seconds for refresh when auto-refresh is enabled     |

Example:

```bash
docker run \
  -p 8080:8080 \
  -e APP_IP_MAP_FILE=/data/zones.csv \
  -e APP_PAC_ROOT=/data/pacs \
  -e APP_CONTACT_INFO="NOC Germany" \
  -e APP_DO_AUTO_REFRESH=true \
  -e APP_MAX_CACHE_AGE=3600 \
  pacserver
```

Migration note: previous versions used a `config.yml`. That file is no longer read. Please set the corresponding `APP_` environment variables instead.

### Zones
Zones map IP Networks to PAC Files
The program expects a CSV, each row is one rule, and it supports the following columns

| Column ID | type | Description                                                                                                                                      |
|-----------|------|--------------------------------------------------------------------------------------------------------------------------------------------------|
| 0         | ip   | The Network Address of this rule                                                                                                                 |
| 1         | int  | The (CIDR) Network Size                                                                                                                          |
| 2         | file | The path to the PAC file to use, relative to `pacRoot`                                                                                           |

### PACs
Lastly, you need to provide the PAC Files themselves.
The application allows for the Use of some Template variables.
The known variables are:

| Variable | Description                                          |
|----------|------------------------------------------------------|
| Filename | The (relative) Filename of th file being server      |
| Contact  | Generic Contact Information provided via env         |

To use them, you can use the following Syntax `{{ .<var name> }}`

Below you can find an example:

```js
// Welcome
// This is the {{ .Filename }} PACfile
// For Changes please reach out to {{ .Contact }}

var proxy = "proxy01:8080"

function FindProxyForURL(url, host) {
    if (host === "localhost"
        || isInNet(host, "127.0.0.0", "255.0.0.0")
    ) {
        return "DIRECT"
    }

    return "PROXY " + proxy
}
```

## Metrics

This service exposes Prometheus metrics at `/metrics`.

- HTTP metrics (per request):
  - `http_requests_total{method,endpoint,status}`
  - `http_request_duration_seconds_bucket|sum|count{method,endpoint,status}`

- TCP connection state metrics (system-wide):
  - `tcp_connection_states{state, family}` — counts current TCP sockets grouped by state and IP family.
    - Examples of `state`: `ESTABLISHED`, `LISTEN`, `TIME_WAIT`, `CLOSE_WAIT`, etc.
    - `family`: `ipv4` or `ipv6`.

Notes:
- In multi-process (Gunicorn) mode, metrics are aggregated using Prometheus's multiprocess mode. The start script sets `PROMETHEUS_MULTIPROC_DIR` and cleans stale files automatically.
- TCP state metrics read from `/proc/net/tcp` and `/proc/net/tcp6`. On non-Linux systems or if `/proc` is unavailable, the metric yields no samples.