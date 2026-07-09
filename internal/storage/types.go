package storage

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/timeforaninja/pacserver/pkg/IP"
)

// StorageConfig contains the file-system inputs the storage package needs.
type StorageConfig struct {
	IPMapFile      string
	PACRoot        string
	DefaultPACFile string
	WPADFile       string
	ContactInfo    string
	IgnoreMinors   bool
}

// IPMap is one parsed row from the zone file.
type IPMap struct {
	IPNet    IP.Net `json:"IPNet"`
	Filename string `json:"Filename"`
	Comment  string `json:"Comment"`
}

// PACTemplate is one PAC file loaded from disk.
type PACTemplate struct {
	Filename string `json:"Filename"`
	Content  string `json:"Content"`
}

// LookupEntry is the paired zone/template record stored in the tree.
type LookupEntry struct {
	IPMap   *IPMap
	PAC     *PACTemplate
	Variant string
}

func (e *LookupEntry) Stringify() string {
	if e == nil || e.IPMap == nil {
		return ""
	}

	comment := ""
	if e.IPMap.Comment != "" {
		comment = fmt.Sprintf(" // %s", e.IPMap.Comment)
	}

	return fmt.Sprintf(
		"%s | pac(%s)%s",
		e.IPMap.IPNet.ToString(),
		e.IPMap.Filename,
		comment,
	)
}

func (e *LookupEntry) IsIdentical(other any) bool {
	if e == nil || e.PAC == nil || other == nil {
		return false
	}

	// Two lookup entries are considered identical when they point at the same PAC file.
	switch o := other.(type) {
	case *LookupEntry:
		return o != nil && o.PAC != nil && e.PAC.Filename == o.PAC.Filename
	case LookupEntry:
		return o.PAC != nil && e.PAC.Filename == o.PAC.Filename
	default:
		return false
	}
}

// templateParams is a struct used to fill in the PAC template.
type templateParams struct {
	Filename string
	Contact  string
}

// NewLookupEntry compiles a PAC template for a zone.
func NewLookupEntry(ipMap *IPMap, pac *PACTemplate, contactInfo string) (*LookupEntry, error) {
	if ipMap == nil || pac == nil {
		return nil, fmt.Errorf("ip map and pac template must be provided")
	}

	// Render the PAC template immediately so later lookups only need to serve the compiled body.
	filledTemplate, err := template.New("pac-template").Parse(pac.Content)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	data := templateParams{Filename: pac.Filename, Contact: contactInfo}
	if err := filledTemplate.Execute(&buf, data); err != nil {
		return nil, err
	}

	return &LookupEntry{
		IPMap:   ipMap,
		PAC:     pac,
		Variant: buf.String(),
	}, nil
}
