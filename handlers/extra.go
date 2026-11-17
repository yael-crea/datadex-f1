package handlers

import (
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strings"
	"time"

	"datadex-f1/api"
	"datadex-f1/collection"
)

// teams : collection des écuries (recherche + filtre nationalité + pagination).
func (a *App) teams(w http.ResponseWriter, r *http.Request) {
	all, err := a.api.AllConstructors()
	if err != nil {
		a.apiError(w, err)
		return
	}
	q := r.URL.Query()
	query := strings.TrimSpace(q.Get("q"))
	nationality := q.Get("nationalite")

	// Liste des nationalités disponibles, construite à la main.
	var nationalities []string
	for _, t := range all {
		if t.Nationality != "" && !slices.Contains(nationalities, t.Nationality) {
			nationalities = append(nationalities, t.Nationality)
		}
	}
	sort.Strings(nationalities)
	if !slices.Contains(nationalities, nationality) {
		nationality = ""
	}

	var filtered []api.Constructor
	for _, t := range collection.SearchConstructors(all, query) {
		if nationality == "" || t.Nationality == nationality {
			filtered = append(filtered, t)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		return collection.Normalize(filtered[i].Name) < collection.Normalize(filtered[j].Name)
	})

	params := url.Values{}
	if query != "" {
		params.Set("q", query)
	}
	if nationality != "" {
		params.Set("nationalite", nationality)
	}
	pageItems, pager := collection.Paginate(filtered,
		collection.ParsePage(q), collection.ParsePerPage(q), "/ecuries", params)

	a.render(w, http.StatusOK, "teams", data{
		"Title":         "Écuries",
		"Active":        "teams",
		"Teams":         pageItems,
		"Total":         len(all),
		"Count":         len(filtered),
		"TeamQuery":     query,
		"Nationality":   nationality,
		"Nationalities": nationalities,
		"Pager":         pager,
		"PagerAction":   "/ecuries",
		"PagerParams":   params,
	})
}

// raceRow est une ligne du calendrier avec son statut.
type raceRow struct {
	api.Race
	Status string // "passe", "prochain" ou "avenir"
}

// calendar : calendrier de la saison en cours.
func (a *App) calendar(w http.ResponseWriter, r *http.Request) {
	season, races, err := a.api.CurrentRaces()
	if err != nil {
		a.apiError(w, err)
		return
	}
	next, hasNext := nextRace(races, time.Now())
	rows := make([]raceRow, len(races))
	done := 0
	for i, race := range races {
		status := "avenir"
		switch {
		case hasNext && race.Round == next.Round:
			status = "prochain"
		case !hasNext || race.Date < next.Date:
			status = "passe"
			done++
		}
		rows[i] = raceRow{Race: race, Status: status}
	}
	a.render(w, http.StatusOK, "calendar", data{
		"Title":  "Calendrier " + season,
		"Active": "calendar",
		"Season": season,
		"Races":  rows,
		"Done":   done,
	})
}

// about : page « à propos » avec la documentation et la FAQ.
func (a *App) about(w http.ResponseWriter, r *http.Request) {
	a.render(w, http.StatusOK, "about", data{"Title": "À propos", "Active": "about"})
}
