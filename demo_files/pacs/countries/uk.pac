// Welcome
// This is the {{ .Filename }} PAC-File
// For Changes please reach out to {{ .Contact }}

var proxy = "uk-proxy:8080"

function FindProxyForURL(url, host) {
    if (host === "localhost"
        || isInNet(host, "127.0.0.0", "255.0.0.0")
    ) {
        return "DIRECT"
    }

    if (shExpMatch(host, "*.gov.uk")) {
        return "DIRECT"
    }

    if (shExpMatch(host, "*.co.uk")) {
        return "PROXY uk-business-proxy:3128"
    }

    return "PROXY " + proxy
}