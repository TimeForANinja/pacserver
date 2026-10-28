package internal

import "testing"

func TestRequestVerdict(t *testing.T) {
	for _, tc := range []struct {
		name     string
		features requestFeatures
		wantIP   string
		wantPath string
	}{
		{
			name:     "explicit route IP wins over the route path",
			features: requestFeatures{RouteIP: "10.1.2.3", RouteNetworkBits: 24, RawRoute: "/10.1.2.3/24", XForwardedForIP: "192.0.2.9", SourceIP: "198.51.100.7"},
			wantIP:   "10.1.2.3",
			wantPath: "",
		},
		{
			name:     "route path carries the X-Forwarded-For IP",
			features: requestFeatures{RawRoute: "/favicon.ico", XForwardedForIP: "192.0.2.9", SourceIP: "198.51.100.7"},
			wantIP:   "192.0.2.9",
			wantPath: "/favicon.ico",
		},
		{
			name:     "route path falls back to the source IP",
			features: requestFeatures{RawRoute: "/favicon.ico", SourceIP: "198.51.100.7"},
			wantIP:   "198.51.100.7",
			wantPath: "/favicon.ico",
		},
		{
			name:     "root path resolves by X-Forwarded-For",
			features: requestFeatures{RawRoute: "/", XForwardedForIP: "192.0.2.9", SourceIP: "198.51.100.7"},
			wantIP:   "192.0.2.9",
			wantPath: "",
		},
		{
			name:     "root path resolves by the source IP",
			features: requestFeatures{RawRoute: "/", SourceIP: "198.51.100.7"},
			wantIP:   "198.51.100.7",
			wantPath: "",
		},
		{
			name:     "no usable IP still returns the route path",
			features: requestFeatures{RawRoute: "/some/unknown/path"},
			wantIP:   "",
			wantPath: "/some/unknown/path",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gotIP, gotPath := requestVerdict(tc.features)
			if gotIP != tc.wantIP || gotPath != tc.wantPath {
				t.Fatalf("requestVerdict = (%q, %q), want (%q, %q)", gotIP, gotPath, tc.wantIP, tc.wantPath)
			}
		})
	}
}
