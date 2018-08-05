package main

import (
	"encoding/json"
	"html/template"
	"os"

	api "github.com/ehazlett/stellar/api/services/application/v1"
)

func appInspectOutputText(app *api.App) error {
	t := template.New("app")
	tmpl, err := t.Parse(appInspectTemplate)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(os.Stdout, app); err != nil {
		return err
	}
	return nil
}

func appInspectOutputJSON(app *api.App) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")
	return enc.Encode(app)
}
