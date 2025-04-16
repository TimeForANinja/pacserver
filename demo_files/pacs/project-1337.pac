// Welcome
// This is the {{ .Filename }} PAC-File
// For Changes please reach out to {{ .Contact }}

var proxy = "my-proxy-31"

function FindProxyForURL(url, host) {
    return "PROXY " + proxy
}
