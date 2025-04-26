// Welcome
// This is the {{ .Filename }} PAC-File
// For Changes please reach out to {{ .Contact }}

var proxy = "china-proxy:8080"

function FindProxyForURL(url, host) {
    if (host === "localhost"
        || isInNet(host, "127.0.0.0", "255.0.0.0")
    ) {
        return "DIRECT"
    }

    if (shExpMatch(host, "*.gov.cn")) {
        return "DIRECT"
    }

    if (shExpMatch(host, "*.com.cn")) {
        return "PROXY china-business-proxy:3128"
    }

    if (shExpMatch(host, "*.cn")) {
        return "PROXY china-regional-proxy:3128"
    }

    return "PROXY " + proxy
}