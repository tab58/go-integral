package seedgen

import (
	"bytes"
	_ "embed"
	"text/template"
)

//go:embed templates/seed_script.tmpl
var seedScriptTemplate string

func GenerateSeedScriptFromTableSchemas(schemas []TableSchema) (string, error) {
	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
	}
	tmpl, err := template.New("seed_script").Funcs(funcMap).Parse(seedScriptTemplate)
	if err != nil {
		return "", err
	}

	buf := bytes.Buffer{}
	err = tmpl.Execute(&buf, schemas)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

//go:embed templates/table_record.tmpl
var tableRecordTemplate string

func GenerateGoFileFromTableSchema(schema TableSchema) (string, error) {
	funcMap := template.FuncMap{
		// The name "inc" is what the function will be called in the template text.
		"inc": func(i int) int {
			return i + 1
		},
	}

	tmpl, err := template.New("table_record").Funcs(funcMap).Parse(tableRecordTemplate)
	if err != nil {
		return "", err
	}

	buf := bytes.Buffer{}
	err = tmpl.Execute(&buf, schema)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
