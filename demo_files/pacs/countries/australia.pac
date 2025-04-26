// Welcome
// This is the {{ .Filename }} PAC-File
// For Changes please reach out to {{ .Contact }}

var proxy = "australia-proxy:8080"

function FindProxyForURL(url, host) {
    if (host === "localhost"
        || isInNet(host, "127.0.0.0", "255.0.0.0")
    ) {
        return "DIRECT"
    }

    if (shExpMatch(host, "*.gov.au")) {
        return "DIRECT"
    }

    if (shExpMatch(host, "*.edu.au")) {
        return "PROXY australia-edu-proxy:3128"
    }

    if (shExpMatch(host, "*.com.au")) {
        return "PROXY australia-business-proxy:3128"
    }

    return "PROXY " + proxy
}