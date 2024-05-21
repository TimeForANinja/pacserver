// Welcome
// This is the {{ .Filename }} PAC-File
// For Changes please reach out to {{ .Contact }}

var proxy = "{{ .Proxy }}"

function FindProxyForURL(url, host) {
    return "PROXY " + proxy
}
