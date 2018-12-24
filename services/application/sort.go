package application

import (
	api "github.com/ehazlett/stellar/api/services/application/v1"
)

// AppSorter sorts applications by name
type AppSorter []*api.App

func (a AppSorter) Len() int           { return len(a) }
func (a AppSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a AppSorter) Less(i, j int) bool { return a[i].Name < a[j].Name }
