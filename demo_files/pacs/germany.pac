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

    if (shExpMatch(host, "*.example.com")) {
        return "PROXY 1.2.3.4:420"
    }

    return "PROXY " + proxy
}
