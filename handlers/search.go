package handlers

import (
	"net/http"
	"strings"

	"datadex-f1/collection"
)

// maxQueryLength évite qu'une recherche trop longue n'occupe le serveur.
const maxQueryLength = 80

// search : page de résultats de recherche (pilotes + écuries).
func (a *App) search(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if len([]rune(query)) > maxQueryLength {
		query = string([]rune(query)[:maxQueryLength])
	}

	d := data{"Title": "Recherche", "Active": "search", "Query": query}
	if query == "" {
		a.render(w, http.StatusOK, "search", d)
		return
	}

	drivers, err := a.api.AllDrivers()
	if err != nil {
		a.apiError(w, err)
		return
	}
	teams, err := a.api.AllConstructors()
	if err != nil {
		a.apiError(w, err)
		return
	}

	foundDrivers := collection.SearchDrivers(drivers, query)
	foundTeams := collection.SearchConstructors(teams, query)

	d["Title"] = "Recherche : " + query
	d["Drivers"] = foundDrivers
	d["DriverCount"] = len(foundDrivers)
	d["Teams"] = foundTeams
	d["TeamCount"] = len(foundTeams)
	d["Searched"] = true
	a.render(w, http.StatusOK, "search", d)
}
