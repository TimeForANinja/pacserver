// Welcome
// This is the {{ .Filename }} PAC-File
// For Changes please reach out to {{ .Contact }}

var proxy = "japan-proxy:8080"

function FindProxyForURL(url, host) {
    if (host === "localhost"
        || isInNet(host, "127.0.0.0", "255.0.0.0")
    ) {
        return "DIRECT"
    }

    if (shExpMatch(host, "*.go.jp")) {
        return "DIRECT"
    }

    if (shExpMatch(host, "*.co.jp")) {
        return "PROXY japan-business-proxy:3128"
    }

    if (shExpMatch(host, "*.jp")) {
        return "PROXY japan-regional-proxy:3128"
    }

    return "PROXY " + proxy
}