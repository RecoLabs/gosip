package fba

import (
	"bytes"
	"strings"
	"text/template"
)

func buildFbaWsTemplate(username, password string) (string, error) {
	type fbaWsTemplate struct {
		Username string
		Password string
	}

	template, err := template.New("fbaWsTemplate").Parse(`
		<?xml version="1.0" encoding="utf-8"?>
		<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
			<soap:Body>
				<Login xmlns="http://schemas.microsoft.com/sharepoint/soap/">
					<username>{{.Username}}</username>
					<password>{{.Password}}</password>
				</Login>
			</soap:Body>
		</soap:Envelope>
	`)
	if err != nil {
		return "", err
	}

	data := fbaWsTemplate{
		Username: escapeParamString(username),
		Password: escapeParamString(password),
	}

	var tpl bytes.Buffer
	if err := template.Execute(&tpl, data); err != nil {
		return "", err
	}

	result := compactTemplate(tpl.String())

	return result, nil
}

func escapeParamString(s string) string {
	s = strings.Replace(s, "&", "&amp;", -1)
	s = strings.Replace(s, "\"", "&quot;", -1)
	s = strings.Replace(s, "'", "&apos;", -1)
	s = strings.Replace(s, "<", "&lt;", -1)
	s = strings.Replace(s, ">", "&gt;", -1)
	return s
}

func compactTemplate(s string) string {
	var result string
	for _, line := range strings.Split(s, "\n") {
		if l := strings.TrimSpace(line); len(l) > 0 {
			result += l
		}
	}
	return result
}
