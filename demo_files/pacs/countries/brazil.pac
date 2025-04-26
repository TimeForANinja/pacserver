// Welcome
// This is the {{ .Filename }} PAC-File
// For Changes please reach out to {{ .Contact }}

var proxy = "brazil-proxy:8080"

function FindProxyForURL(url, host) {
    if (host === "localhost"
        || isInNet(host, "127.0.0.0", "255.0.0.0")
    ) {
        return "DIRECT"
    }

    if (shExpMatch(host, "*.gov.br")) {
        return "DIRECT"
    }

    if (shExpMatch(host, "*.com.br")) {
        return "PROXY brazil-business-proxy:3128"
    }

    if (shExpMatch(host, "*.br")) {
        return "PROXY brazil-regional-proxy:3128"
    }

    return "PROXY " + proxy
}