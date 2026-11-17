// Package handlers contient le serveur web : routes, pages et erreurs.
package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"datadex-f1/api"
	"datadex-f1/favorites"
)

// App regroupe les dépendances partagées par toutes les pages.
type App struct {
	api       *api.Client
	favs      *favorites.Store
	templates map[string]*template.Template
}

// data est le contenu envoyé à un template.
type data map[string]any

// pages liste les templates : chacun est combiné avec layout.html.
var pages = []string{"home", "drivers", "driver", "search", "favorites", "teams", "calendar", "about", "error"}

// New charge tous les templates et prépare l'application.
func New(client *api.Client, favs *favorites.Store, templateDir string) (*App, error) {
	app := &App{api: client, favs: favs, templates: make(map[string]*template.Template)}
	for _, name := range pages {
		t, err := template.New("layout.html").Funcs(templateFuncs).ParseFiles(
			filepath.Join(templateDir, "layout.html"),
			filepath.Join(templateDir, name+".html"),
		)
		if err != nil {
			return nil, fmt.Errorf("template %s : %w", name, err)
		}
		app.templates[name] = t
	}
	return app, nil
}

// Routes déclare toutes les URL du site.
func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("GET /{$}", a.home)
	mux.HandleFunc("GET /pilotes", a.drivers)
	mux.HandleFunc("GET /pilotes/{id}", a.driverDetail)
	mux.HandleFunc("GET /recherche", a.search)
	mux.HandleFunc("GET /ecuries", a.teams)
	mux.HandleFunc("GET /calendrier", a.calendar)
	mux.HandleFunc("GET /a-propos", a.about)
	mux.HandleFunc("GET /favoris", a.favoritesPage)
	mux.HandleFunc("POST /favoris/ajouter", a.addFavorite)
	mux.HandleFunc("POST /favoris/supprimer", a.removeFavorite)
	mux.HandleFunc("POST /favoris/vider", a.clearFavorites)

	// Toute autre URL : page 404 personnalisée.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		a.notFound(w, "Cette page n'existe pas… elle a dû rester au stand.")
	})
	return logRequests(recoverPanics(a, mux))
}

// render exécute un template dans un tampon avant d'écrire la réponse :
// si le template plante, on peut encore envoyer une vraie page d'erreur.
func (a *App) render(w http.ResponseWriter, status int, name string, d data) {
	t, ok := a.templates[name]
	if !ok {
		http.Error(w, "template introuvable", http.StatusInternalServerError)
		return
	}
	if d == nil {
		d = data{}
	}
	d["Year"] = time.Now().Year()
	d["FavCount"] = len(a.favs.IDs())
	for _, key := range []string{"Active", "Query"} {
		if _, ok := d[key]; !ok {
			d[key] = ""
		}
	}

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "layout", d); err != nil {
		log.Printf("erreur template %s : %v", name, err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

// errorPage affiche la page d'erreur générique.
func (a *App) errorPage(w http.ResponseWriter, status int, title, message string) {
	a.render(w, status, "error", data{
		"Title":   title,
		"Status":  status,
		"Heading": title,
		"Message": message,
	})
}

func (a *App) notFound(w http.ResponseWriter, message string) {
	a.errorPage(w, http.StatusNotFound, "Page introuvable", message)
}

// apiError traduit une erreur du client API en page d'erreur adaptée.
func (a *App) apiError(w http.ResponseWriter, err error) {
	log.Printf("erreur API : %v", err)
	switch {
	case errors.Is(err, api.ErrNotFound):
		a.notFound(w, "Ce pilote ou cette ressource n'existe pas dans les archives de la F1.")
	case errors.Is(err, api.ErrRateLimited):
		a.errorPage(w, http.StatusServiceUnavailable, "Drapeau jaune",
			"L'API F1 reçoit trop de demandes. Patientez quelques secondes puis rechargez la page.")
	default:
		a.errorPage(w, http.StatusBadGateway, "Drapeau rouge : API hors ligne",
			"Impossible de joindre l'API Jolpica-F1 pour le moment. Vérifiez votre connexion ou réessayez plus tard.")
	}
}

// recoverPanics empêche le serveur de s'arrêter si une page plante.
func recoverPanics(a *App, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic sur %s : %v", r.URL.Path, rec)
				a.errorPage(w, http.StatusInternalServerError, "Erreur interne",
					"Un problème inattendu est survenu. L'équipe est déjà dans le garage.")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.RequestURI(), time.Since(start).Round(time.Millisecond))
	})
}

// driverCard est un pilote enrichi des infos nécessaires au bouton favori.
type driverCard struct {
	api.Driver
	IsFav  bool
	Return string // page où revenir après l'ajout/la suppression
}

// cards prépare les cartes des pilotes affichés sur la page courante.
func (a *App) cards(drivers []api.Driver, r *http.Request) []driverCard {
	favIDs := a.favs.IDs()
	out := make([]driverCard, len(drivers))
	for i, d := range drivers {
		out[i] = driverCard{Driver: d, IsFav: favIDs[d.DriverID], Return: r.URL.RequestURI()}
	}
	return out
}
