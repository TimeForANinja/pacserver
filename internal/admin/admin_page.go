package admin

import (
	"bytes"
	"embed"
	"html/template"
)

//go:embed admin_page.html
//go:embed admin_login.html
var adminPageFS embed.FS

type adminPageData struct {
	PrometheusPath string
}

var adminPageTemplate = template.Must(template.New("admin_page.html").ParseFS(adminPageFS, "admin_page.html"))
var adminLoginTemplate = template.Must(template.New("admin_login.html").ParseFS(adminPageFS, "admin_login.html"))

func renderAdminPage(prometheusPath string) (string, error) {
	var out bytes.Buffer
	if err := adminPageTemplate.Execute(&out, adminPageData{
		PrometheusPath: prometheusPath,
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
