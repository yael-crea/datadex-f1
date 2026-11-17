package handlers

import (
	"log"
	"net/http"
	"time"

	"datadex-f1/api"
)

// home : page d'accueil, présente l'univers F1 et la saison en cours.
func (a *App) home(w http.ResponseWriter, r *http.Request) {
	d := data{"Title": "Accueil", "Active": "home"}

	// L'accueil reste affichable même si une partie des données manque.
	if drivers, err := a.api.AllDrivers(); err == nil {
		d["DriverCount"] = len(drivers)
	}
	if teams, err := a.api.AllConstructors(); err == nil {
		d["TeamCount"] = len(teams)
	}

	season, round, standings, err := a.api.CurrentStandings()
	if err != nil {
		log.Printf("accueil : classement indisponible (%v)", err)
		d["StandingsError"] = true
	} else {
		if len(standings) > 10 {
			standings = standings[:10]
		}
		d["Season"], d["Round"], d["Standings"] = season, round, standings
	}

	if _, races, err := a.api.CurrentRaces(); err == nil {
		if next, ok := nextRace(races, time.Now()); ok {
			d["NextRace"] = next
		}
		d["RaceCount"] = len(races)
	}

	a.render(w, http.StatusOK, "home", d)
}

// nextRace renvoie le premier Grand Prix qui n'a pas encore eu lieu.
func nextRace(races []api.Race, now time.Time) (api.Race, bool) {
	today := now.Truncate(24 * time.Hour)
	for _, race := range races {
		if day, ok := race.Day(); ok && !day.Before(today) {
			return race, true
		}
	}
	return api.Race{}, false
}

// drivers : collection de tous les pilotes.
func (a *App) drivers(w http.ResponseWriter, r *http.Request) {
	all, err := a.api.AllDrivers()
	if err != nil {
		a.apiError(w, err)
		return
	}
	a.render(w, http.StatusOK, "drivers", data{
		"Title":   "Pilotes",
		"Active":  "drivers",
		"Drivers": all,
		"Total":   len(all),
	})
}

// driverStats regroupe les chiffres affichés sur la fiche d'un pilote.
type driverStats struct {
	Races   int
	Wins    int
	Seasons []api.Season
	Teams   []api.Constructor
	Partial bool // vrai si une partie des statistiques n'a pas pu être chargée
}

// FirstSeason et LastSeason servent à afficher la période de carrière.
func (s driverStats) FirstSeason() string {
	if len(s.Seasons) == 0 {
		return "—"
	}
	return s.Seasons[0].Season
}

func (s driverStats) LastSeason() string {
	if len(s.Seasons) == 0 {
		return "—"
	}
	return s.Seasons[len(s.Seasons)-1].Season
}

// WinRate renvoie le pourcentage de victoires arrondi.
func (s driverStats) WinRate() int {
	if s.Races == 0 {
		return 0
	}
	return s.Wins * 100 / s.Races
}

// driverDetail : fiche détaillée d'un pilote.
func (a *App) driverDetail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	driver, err := a.api.Driver(id)
	if err != nil {
		a.apiError(w, err)
		return
	}

	var stats driverStats
	if stats.Seasons, err = a.api.DriverSeasons(id); err != nil {
		stats.Partial = true
	}
	if stats.Teams, err = a.api.DriverConstructors(id); err != nil {
		stats.Partial = true
	}
	if stats.Races, err = a.api.DriverRaceCount(id); err != nil {
		stats.Partial = true
	}
	if stats.Wins, err = a.api.DriverWins(id); err != nil {
		stats.Partial = true
	}

	a.render(w, http.StatusOK, "driver", data{
		"Title":  driver.FullName(),
		"Active": "drivers",
		"Driver": driver,
		"Stats":  stats,
	})
}
