# PAC-Server
This is a custom server made to serve Proxy Auto Config (short PAC) Files based on Source IP.

## Setup
The setup of this App is pretty straight forward.
After downloading the executable you only require the following three parameters:

### Config
The application expects a `./config.yml` in the cwd.
The supported fields for that yaml are:

| Field         | Type   | Description                                                     |
|---------------|--------|-----------------------------------------------------------------|
| ipMapFile     | string | path to the Zones `.csv` file                                   |
| pacRoot       | string | path to the directory containing the PAC Files                  |
| contactInfo   | string | Contact Info that can be used inside the PAC Templates          |
| accessLogFile | string | the path to the access log file                                 |
| eventLogFile  | string | the path to the event log file                                  |
| doAutoRefresh | bool   | Yes to Automatically reload PAC and Zones in a regular interval |
| maxCacheAge   | int    | The interval (in seconds) to reload the PAC and Zone files in   |
| prometheusEnabled | bool | Enable Prometheus metrics collection and exposure |
| prometheusPath | string | The endpoint path for exposing Prometheus metrics (default: "/metrics") |

### Zones
Zones map IP Networks to PAC Files
The program expects a CSV, each row is one rule and it supports the following columns

| Column ID | type | Description                                                                                                                                      |
|-----------|------|--------------------------------------------------------------------------------------------------------------------------------------------------|
| 0         | ip   | The Network Address of this rule                                                                                                                 |
| 1         | int  | The (CIDR) Network Size                                                                                                                          |
| 2         | file | The path to the PAC file to use, relative to `pacRoot`                                                                                           |

### PACs
Lastly you need to provide the PAC Files themselves.
The application allows for the Use of some Template variables.
The known variables are:

| Variable | Description                                          |
|----------|------------------------------------------------------|
| Filename | The (relative) Filename of th file being server      |
| Contact  | Generic Contact Information provided in `config.yml` |

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

## Prometheus Metrics

The PAC-Server includes built-in support for Prometheus metrics to monitor performance. When enabled, the server exposes various metrics that can be scraped by Prometheus and visualized in Grafana.

### Enabling Prometheus Metrics

To enable Prometheus metrics, set the following in your `config.yml`:

```yaml
prometheusEnabled: true
prometheusPath: "/metrics"  # The endpoint where metrics will be exposed
```

### Available Metrics

The following metrics are available:

#### Request/Response Metrics
- **Request Rate**: Average and peak requests per second
  - `http_requests_total` - Total number of HTTP requests
  - `http_request_duration_seconds` - HTTP request latency in seconds

- **Response Time**: Average and peak response times
  - `pacserver_response_time_seconds` - Response time distribution in seconds

- **Data I/O**: Bytes transferred
  - `pacserver_data_in_bytes_total` - Total bytes received
  - `pacserver_data_out_bytes_total` - Total bytes sent

- **HTTP Error Rate**: Count of HTTP errors by status code
  - `pacserver_http_errors_total` - Total number of HTTP errors

#### System Metrics
- **Resource Utilization**:
  - `pacserver_memory_usage_bytes` - Current memory usage in bytes
  - `pacserver_cpu_usage_percent` - Current CPU usage percentage (requires additional setup)

- **Thread Count**:
  - `pacserver_thread_count` - Current number of goroutines

#### Connection Metrics
- Basic connection metrics are available through the Fiber Prometheus middleware

### Grafana Integration

To visualize these metrics in Grafana:

1. Configure Prometheus to scrape the PAC-Server metrics endpoint
2. Add Prometheus as a data source in Grafana
3. Create dashboards using the metrics listed above

Example Prometheus configuration:

```yaml
scrape_configs:
  - job_name: 'pacserver'
    scrape_interval: 15s
    static_configs:
      - targets: ['pacserver:8081']
```
