// Welcome
// This is the {{ .Filename }} PAC-File
// For Changes please reach out to {{ .Contact }}

var proxy = "canada-proxy:8080"

function FindProxyForURL(url, host) {
    if (host === "localhost"
        || isInNet(host, "127.0.0.0", "255.0.0.0")
    ) {
        return "DIRECT"
    }

    if (shExpMatch(host, "*.gc.ca")) {
        return "DIRECT"
    }

    if (shExpMatch(host, "*.ca")) {
        return "PROXY canada-regional-proxy:3128"
    }

    return "PROXY " + proxy
}