package internal

/**
 * LookupElement is the core data structure of this server
 * it maps an IP Net (CIDR) to a PAC File
 */

import (
	"bytes"
	"fmt"
	"html/template"
)

type LookupElement struct {
	IPMap *ipMap       `json:"IPMap"`
	PAC   *pacTemplate `json:"PAC"`
	// the parsed content of the PAC Template
	variant string
}

func (le1 LookupElement) isIdenticalNet(le2 LookupElement) bool {
	return le1.IPMap.IPNet.IsIdentical(le2.IPMap.IPNet)
}

func (le1 LookupElement) isIdenticalPAC(le2 LookupElement) bool {
	// PAC1 can be undefined in testing scenarios
	if le1.PAC == nil || le2.PAC == nil {
		return false
	}
	return le1.PAC.Filename == le2.PAC.Filename
}

func (le1 LookupElement) isSubnetOf(le2 LookupElement) bool {
	return le1.IPMap.IPNet.IsSubnetOf(le2.IPMap.IPNet)
}

func (le1 LookupElement) getRawCIDR() uint8 {
	return le1.IPMap.IPNet.GetRawCIDR()
}

func (le1 LookupElement) getVariant() string {
	return le1.variant
}

func (le1 LookupElement) _stringify() string {
	return fmt.Sprintf(
		"%s | pac(%s)",
		le1.IPMap.IPNet.ToString(),
		le1.IPMap.Filename,
	)
}

type templateParams struct {
	Filename string
	Contact  string
}

func NewLookupElement(ipMap *ipMap, pac *pacTemplate, contactInfo string) (LookupElement, error) {
	pacTemplate, err := template.New("pac-template").Parse(pac.content)
	if err != nil {
		return LookupElement{}, err
	}

	var buf bytes.Buffer
	data := templateParams{pac.Filename, contactInfo}

	err = pacTemplate.Execute(&buf, data)
	if err != nil {
		return LookupElement{}, err
	}

	variant := buf.String()

	return LookupElement{
		IPMap:   ipMap,
		PAC:     pac,
		variant: variant,
	}, nil
}
