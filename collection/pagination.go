package collection

import (
	"net/url"
	"strconv"
)

// PerPageOptions : nombres d'éléments par page proposés à l'utilisateur.
var PerPageOptions = []int{10, 20, 30, 50}

// DefaultPerPage est utilisé si l'utilisateur n'a rien choisi.
const DefaultPerPage = 20

// PageLink est un bouton de la barre de pagination.
type PageLink struct {
	Number   int
	URL      string
	Current  bool
	Ellipsis bool // « … » entre deux groupes de pages
}

// Pagination contient tout ce dont le template a besoin.
type Pagination struct {
	Page       int
	PerPage    int
	TotalItems int
	TotalPages int
	From, To   int // position des éléments affichés (ex : 21 à 40)
	PrevURL    string
	NextURL    string
	Links      []PageLink
	Options    []int
}

// ParsePage lit le numéro de page ; une valeur invalide renvoie 1.
func ParsePage(q url.Values) int {
	p, err := strconv.Atoi(q.Get("page"))
	if err != nil || p < 1 {
		return 1
	}
	return p
}

// ParsePerPage lit le nombre d'éléments par page ; seules les valeurs
// proposées dans PerPageOptions sont acceptées.
func ParsePerPage(q url.Values) int {
	n, err := strconv.Atoi(q.Get("parpage"))
	if err != nil {
		return DefaultPerPage
	}
	for _, opt := range PerPageOptions {
		if n == opt {
			return n
		}
	}
	return DefaultPerPage
}

// Paginate découpe la liste à la main et prépare les liens de navigation.
// basePath + params servent à construire les URL des autres pages en
// conservant la recherche et les filtres de l'utilisateur.
func Paginate[T any](items []T, page, perPage int, basePath string, params url.Values) ([]T, Pagination) {
	total := len(items)
	totalPages := (total + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}
	// Une page trop grande (ex : ?page=999) renvoie la dernière page.
	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * perPage
	end := start + perPage
	if end > total {
		end = total
	}

	p := Pagination{
		Page:       page,
		PerPage:    perPage,
		TotalItems: total,
		TotalPages: totalPages,
		Options:    PerPageOptions,
	}
	if total > 0 {
		p.From, p.To = start+1, end
	}

	link := func(n int) string {
		v := url.Values{}
		for k, vals := range params {
			v[k] = vals
		}
		if n > 1 {
			v.Set("page", strconv.Itoa(n))
		}
		if perPage != DefaultPerPage {
			v.Set("parpage", strconv.Itoa(perPage))
		}
		if len(v) == 0 {
			return basePath
		}
		return basePath + "?" + v.Encode()
	}

	if page > 1 {
		p.PrevURL = link(page - 1)
	}
	if page < totalPages {
		p.NextURL = link(page + 1)
	}

	// On affiche la première, la dernière et 2 pages autour de la page
	// courante : 1 … 4 5 [6] 7 8 … 45
	lastShown := 0
	for n := 1; n <= totalPages; n++ {
		if n == 1 || n == totalPages || (n >= page-2 && n <= page+2) {
			if lastShown != 0 && n-lastShown > 1 {
				p.Links = append(p.Links, PageLink{Ellipsis: true})
			}
			p.Links = append(p.Links, PageLink{Number: n, URL: link(n), Current: n == page})
			lastShown = n
		}
	}

	return items[start:end], p
}
