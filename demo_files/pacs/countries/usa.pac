// Welcome
// This is the {{ .Filename }} PAC-File
// For Changes please reach out to {{ .Contact }}

var proxy = "usa-proxy:8080"

function FindProxyForURL(url, host) {
    if (host === "localhost"
        || isInNet(host, "127.0.0.0", "255.0.0.0")
    ) {
        return "DIRECT"
    }

    if (shExpMatch(host, "*.internal.example.com")) {
        return "DIRECT"
    }

    if (shExpMatch(host, "*.example.com")) {
        return "PROXY usa-special-proxy:3128"
    }

    return "PROXY " + proxy
}