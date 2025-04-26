// Welcome
// This is the {{ .Filename }} PAC-File
// For Changes please reach out to {{ .Contact }}

var proxy = "russia-proxy:8080"

function FindProxyForURL(url, host) {
    if (host === "localhost"
        || isInNet(host, "127.0.0.0", "255.0.0.0")
    ) {
        return "DIRECT"
    }

    if (shExpMatch(host, "*.gov.ru")) {
        return "DIRECT"
    }

    if (shExpMatch(host, "*.com.ru")) {
        return "PROXY russia-business-proxy:3128"
    }

    if (shExpMatch(host, "*.ru")) {
        return "PROXY russia-regional-proxy:3128"
    }

    return "PROXY " + proxy
}