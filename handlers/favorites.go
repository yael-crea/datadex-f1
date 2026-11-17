package handlers

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

	"datadex-f1/api"
	"datadex-f1/favorites"
)

// favoritesPage : affiche et permet de gérer la liste des favoris.
func (a *App) favoritesPage(w http.ResponseWriter, r *http.Request) {
	a.render(w, http.StatusOK, "favorites", data{
		"Title":     "Favoris",
		"Active":    "favorites",
		"Favorites": a.favs.All(),
		"Message":   flashMessages[r.URL.Query().Get("msg")],
	})
}

// Messages affichés après une action (passés dans l'URL après redirection).
var flashMessages = map[string]string{
	"ajoute":   "Pilote ajouté à vos favoris.",
	"retire":   "Pilote retiré de vos favoris.",
	"vide":     "Votre liste de favoris a été vidée.",
	"existant": "Ce pilote était déjà dans vos favoris.",
}

// addFavorite : POST /favoris/ajouter (champ "id").
func (a *App) addFavorite(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.FormValue("id"))
	if !api.ValidID(id) {
		a.errorPage(w, http.StatusBadRequest, "Requête invalide", "L'identifiant du pilote est invalide.")
		return
	}

	// On vérifie que le pilote existe vraiment avant de l'enregistrer.
	driver, err := a.api.Driver(id)
	if err != nil {
		a.apiError(w, err)
		return
	}

	msg := "ajoute"
	err = a.favs.Add(favorites.Favorite{
		ID:          driver.DriverID,
		GivenName:   driver.GivenName,
		FamilyName:  driver.FamilyName,
		Code:        driver.Code,
		Number:      driver.PermanentNumber,
		Nationality: driver.Nationality,
		DateOfBirth: driver.DateOfBirth,
	})
	switch {
	case errors.Is(err, favorites.ErrAlreadyExists):
		msg = "existant"
	case err != nil:
		log.Printf("favoris : %v", err)
		a.errorPage(w, http.StatusInternalServerError, "Sauvegarde impossible",
			"Le fichier des favoris n'a pas pu être enregistré.")
		return
	}
	redirectBack(w, r, msg)
}

// removeFavorite : POST /favoris/supprimer (champ "id").
func (a *App) removeFavorite(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.FormValue("id"))
	err := a.favs.Remove(id)
	if err != nil && !errors.Is(err, favorites.ErrNotFound) {
		log.Printf("favoris : %v", err)
		a.errorPage(w, http.StatusInternalServerError, "Sauvegarde impossible",
			"Le fichier des favoris n'a pas pu être enregistré.")
		return
	}
	redirectBack(w, r, "retire")
}

// clearFavorites : POST /favoris/vider.
func (a *App) clearFavorites(w http.ResponseWriter, r *http.Request) {
	if err := a.favs.Clear(); err != nil {
		log.Printf("favoris : %v", err)
		a.errorPage(w, http.StatusInternalServerError, "Sauvegarde impossible",
			"Le fichier des favoris n'a pas pu être enregistré.")
		return
	}
	http.Redirect(w, r, "/favoris?msg=vide", http.StatusSeeOther)
}

// redirectBack renvoie l'utilisateur sur la page d'où il vient (champ
// "retour"). Seules les adresses internes au site sont acceptées.
func redirectBack(w http.ResponseWriter, r *http.Request, msg string) {
	target := r.FormValue("retour")
	if !strings.HasPrefix(target, "/") || strings.HasPrefix(target, "//") || strings.Contains(target, "\\") {
		target = "/favoris"
	}
	if target == "/favoris" || strings.HasPrefix(target, "/favoris?") {
		target = "/favoris?msg=" + url.QueryEscape(msg)
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}
