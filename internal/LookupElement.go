package internal

import (
	"bytes"
	"html/template"

	"github.com/timeforaninja/pacserver/pkg/IP"
	"github.com/timeforaninja/pacserver/pkg/utils"
)

type lookupElement struct {
	IPMap *ipMap       `json:"IPMap"`
	PAC   *pacTemplate `json:"PAC"`

	variants []string
}

func (le1 lookupElement) isIdenticalNet(le2 lookupElement) bool {
	return le1.IPMap.IPNet.IsIdentical(le2.IPMap.IPNet)
}

func (le1 lookupElement) isIdenticalPAC(le2 lookupElement) bool {
	return le1.IPMap.Filename == le2.IPMap.Filename && utils.SlicesEqual(le1.IPMap.Hostnames, le2.IPMap.Hostnames)
}

func (le1 lookupElement) isSubnetOf(le2 lookupElement) bool {
	return le1.IPMap.IPNet.IsSubnetOf(le2.IPMap.IPNet)
}

func (le lookupElement) getRawCIDR() uint8 {
	return le.IPMap.IPNet.GetRawCIDR()
}

func (le lookupElement) getVariant(ip IP.IP) string {
	return le.variants[int(ip.Value)%len(le.variants)]
}

type templateParams struct {
	Filename string
	Proxy    string
	Contact  string
}

func NewLookupElement(ipmap *ipMap, pac *pacTemplate, contactInfo string) (lookupElement, error) {
	variants := make([]string, len(ipmap.Hostnames))

	template, err := template.New("pac-template").Parse(pac.content)
	if err != nil {
		return lookupElement{}, err
	}

	for idx := range variants {
		var buf bytes.Buffer
		data := templateParams{pac.Filename, ipmap.Hostnames[idx], contactInfo}

		err := template.Execute(&buf, data)
		if err != nil {
			return lookupElement{}, err
		}

		variants[idx] = buf.String()
	}

	return lookupElement{
		IPMap:    ipmap,
		PAC:      pac,
		variants: variants,
	}, nil
}
