package internal

import (
	"bytes"
	"html/template"

	"github.com/timeforaninja/pacserver/pkg/IP"
	"github.com/timeforaninja/pacserver/pkg/utils"
)

type LookupElement struct {
	IPMap *ipMap       `json:"IPMap"`
	PAC   *pacTemplate `json:"PAC"`

	variants []string
}

func (le1 LookupElement) isIdenticalNet(le2 LookupElement) bool {
	return le1.IPMap.IPNet.IsIdentical(le2.IPMap.IPNet)
}

func (le1 LookupElement) isIdenticalPAC(le2 LookupElement) bool {
	// PAC1 can be undefined in testing scenarios
	if le1.PAC == nil || le2.PAC == nil {
		return false
	}
	return le1.PAC.Filename == le2.PAC.Filename && utils.SlicesEqual(le1.IPMap.Hostnames, le2.IPMap.Hostnames)
}

func (le1 LookupElement) isSubnetOf(le2 LookupElement) bool {
	return le1.IPMap.IPNet.IsSubnetOf(le2.IPMap.IPNet)
}

func (le1 LookupElement) getRawCIDR() uint8 {
	return le1.IPMap.IPNet.GetRawCIDR()
}

func (le1 LookupElement) getVariant(ip IP.IP) string {
	return le1.variants[int(ip.Value)%len(le1.variants)]
}

type templateParams struct {
	Filename string
	Proxy    string
	Contact  string
}

func NewLookupElement(ipMap *ipMap, pac *pacTemplate, contactInfo string) (LookupElement, error) {
	variants := make([]string, len(ipMap.Hostnames))

	pacTemplate, err := template.New("pac-template").Parse(pac.content)
	if err != nil {
		return LookupElement{}, err
	}

	for idx := range variants {
		var buf bytes.Buffer
		data := templateParams{pac.Filename, ipMap.Hostnames[idx], contactInfo}

		err := pacTemplate.Execute(&buf, data)
		if err != nil {
			return LookupElement{}, err
		}

		variants[idx] = buf.String()
	}

	return LookupElement{
		IPMap:    ipMap,
		PAC:      pac,
		variants: variants,
	}, nil
}
