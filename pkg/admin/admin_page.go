package admin

import (
	"bytes"
	"embed"
	"html/template"
	"strconv"
)

//go:embed admin_page.html
//go:embed admin_login.html
var adminPageFS embed.FS

type adminPageData struct {
	PrometheusPath string
	ServerPort     string
}

var adminPageTemplate = template.Must(template.New("admin_page.html").ParseFS(adminPageFS, "admin_page.html"))
var adminLoginTemplate = template.Must(template.New("admin_login.html").ParseFS(adminPageFS, "admin_login.html"))

func renderAdminPage(prometheusPath string, serverPort uint16) (string, error) {
	var out bytes.Buffer
	if err := adminPageTemplate.Execute(&out, adminPageData{
		PrometheusPath: prometheusPath,
		ServerPort:     strconv.FormatUint(uint64(serverPort), 10),
	}); err != nil {
		return "", err
	}
	return out.String(), nil
}

func renderAdminLoginPage() (string, error) {
	var out bytes.Buffer
	if err := adminLoginTemplate.Execute(&out, nil); err != nil {
		return "", err
	}
	return out.String(), nil
}
