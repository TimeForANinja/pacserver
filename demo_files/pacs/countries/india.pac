// Welcome
// This is the {{ .Filename }} PAC-File
// For Changes please reach out to {{ .Contact }}

var proxy = "india-proxy:8080"

function FindProxyForURL(url, host) {
    if (host === "localhost"
        || isInNet(host, "127.0.0.0", "255.0.0.0")
    ) {
        return "DIRECT"
    }

    if (shExpMatch(host, "*.gov.in")) {
        return "DIRECT"
    }

    if (shExpMatch(host, "*.co.in")) {
        return "PROXY india-business-proxy:3128"
    }

    if (shExpMatch(host, "*.in")) {
        return "PROXY india-regional-proxy:3128"
    }

    return "PROXY " + proxy
}