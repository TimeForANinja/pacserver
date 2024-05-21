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

### Zones
Zones map IP Networks to PAC Files
The program expects a CSV, each row is one rule and it supports the following columns

| Column ID | type | Description                                                                                                                                      |
|-----------|------|--------------------------------------------------------------------------------------------------------------------------------------------------|
| 0         | ip   | The Network Address of this rule                                                                                                                 |
| 1         | int  | The (CIDR) Network Size                                                                                                                          |
| 2         | file | The path to the PAC file to use, relative to `pacRoot`                                                                                           |
| 3-        | host | Variables to use for Proxies. If you provide more than one column it will use them as alternative Proxies and do Source IP Based load balancing. |

### PACs
Lastly you need to provide the PAC Files themselves.
The application allows for the Use of some Template variables.
The known variables are:

| Variable | Description                                          |
|----------|------------------------------------------------------|
| Filename | The (relative) Filename of th file being server      |
| Proxy    | (One of) the Proxies provided in the `zones.csv`     |
| Contact  | Generic Contact Information provided in `config.yml` |

To use them, you can use the following Syntax `{{ .<var name> }}`

Below you can find an example:

```js
// Welcome
// This is the {{ .Filename }} PACfile
// For Changes please reach out to {{ .Contact }}

var proxy = "{{ .Proxy }}"

function FindProxyForURL(url, host) {
    if (host === "localhost"
        || isInNet(host, "127.0.0.0", "255.0.0.0")
    ) {
        return "DIRECT"
    }

    return "PROXY " + proxy
}
```