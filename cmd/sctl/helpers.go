package main

import (
	"encoding/json"
	"html/template"
	"os"
	"sort"

	api "github.com/ehazlett/stellar/api/services/application/v1"
	runtimeapi "github.com/ehazlett/stellar/api/services/runtime/v1"
)

type ServiceSorter []*runtimeapi.Service

func (s ServiceSorter) Len() int           { return len(s) }
func (s ServiceSorter) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ServiceSorter) Less(i, j int) bool { return s[i].Name < s[j].Name }

func appInspectOutputText(app *api.App) error {
	sort.Sort(ServiceSorter(app.Services))
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
	sort.Sort(ServiceSorter(app.Services))
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")
	return enc.Encode(app)
}
