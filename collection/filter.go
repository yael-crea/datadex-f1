package collection

import (
	"net/url"
	"sort"
	"strconv"
	"strings"

	"datadex-f1/api"
)

// DriverFilters contient les filtres choisis par l'utilisateur.
// Ils sont tous cumulables : un pilote doit respecter TOUS les filtres actifs.
type DriverFilters struct {
	Query       string // recherche texte (FT1)
	Nationality string // filtre 1 : nationalité
	Decade      int    // filtre 2 : décennie de naissance (0 = toutes)
	Number      string // filtre 3 : "avec" / "sans" numéro permanent
	Sort        string // tri : "nom", "recent", "ancien", "numero"
}

// Valeurs autorisées pour les listes déroulantes.
var (
	NumberOptions = []string{"avec", "sans"}
	SortOptions   = []string{"nom", "recent", "ancien", "numero"}
)

// ParseDriverFilters lit les filtres depuis l'URL. Une valeur invalide
// est simplement ignorée : le site ne plante jamais à cause d'un paramètre.
func ParseDriverFilters(q url.Values) DriverFilters {
	f := DriverFilters{
		Query:       strings.TrimSpace(q.Get("q")),
		Nationality: strings.TrimSpace(q.Get("nationalite")),
		Number:      q.Get("numero"),
		Sort:        q.Get("tri"),
	}
	if len([]rune(f.Query)) > 80 {
		f.Query = string([]rune(f.Query)[:80])
	}
	if dec, err := strconv.Atoi(q.Get("decennie")); err == nil && dec >= 1890 && dec <= 2020 && dec%10 == 0 {
		f.Decade = dec
	}
	if !contains(NumberOptions, f.Number) {
		f.Number = ""
	}
	if !contains(SortOptions, f.Sort) {
		f.Sort = "nom"
	}
	return f
}

// Active indique si au moins un filtre est appliqué.
func (f DriverFilters) Active() bool {
	return f.Query != "" || f.Nationality != "" || f.Decade != 0 || f.Number != ""
}

// Values reconstruit les paramètres d'URL (utilisé par la pagination
// pour conserver les filtres quand on change de page).
func (f DriverFilters) Values() url.Values {
	v := url.Values{}
	if f.Query != "" {
		v.Set("q", f.Query)
	}
	if f.Nationality != "" {
		v.Set("nationalite", f.Nationality)
	}
	if f.Decade != 0 {
		v.Set("decennie", strconv.Itoa(f.Decade))
	}
	if f.Number != "" {
		v.Set("numero", f.Number)
	}
	if f.Sort != "" && f.Sort != "nom" {
		v.Set("tri", f.Sort)
	}
	return v
}

// FilterDrivers parcourt la collection et ne garde que les pilotes
// correspondant à tous les filtres actifs, puis applique le tri.
func FilterDrivers(drivers []api.Driver, f DriverFilters) []api.Driver {
	result := SearchDrivers(drivers, f.Query)

	var out []api.Driver
	for _, d := range result {
		if f.Nationality != "" && d.Nationality != f.Nationality {
			continue
		}
		if f.Decade != 0 && d.Decade() != f.Decade {
			continue
		}
		if f.Number == "avec" && d.PermanentNumber == "" {
			continue
		}
		if f.Number == "sans" && d.PermanentNumber != "" {
			continue
		}
		out = append(out, d)
	}

	SortDrivers(out, f.Sort)
	return out
}

// SortDrivers trie les pilotes selon le critère demandé.
func SortDrivers(drivers []api.Driver, by string) {
	sort.SliceStable(drivers, func(i, j int) bool {
		a, b := drivers[i], drivers[j]
		switch by {
		case "recent":
			return a.DateOfBirth > b.DateOfBirth
		case "ancien":
			return a.DateOfBirth < b.DateOfBirth
		case "numero":
			na, errA := strconv.Atoi(a.PermanentNumber)
			nb, errB := strconv.Atoi(b.PermanentNumber)
			if errA != nil {
				return false // les pilotes sans numéro vont à la fin
			}
			if errB != nil {
				return true
			}
			return na < nb
		default:
			if Normalize(a.FamilyName) == Normalize(b.FamilyName) {
				return Normalize(a.GivenName) < Normalize(b.GivenName)
			}
			return Normalize(a.FamilyName) < Normalize(b.FamilyName)
		}
	})
}

// Nationalities renvoie la liste unique des nationalités présentes.
func Nationalities(drivers []api.Driver) []string {
	seen := map[string]bool{}
	var out []string
	for _, d := range drivers {
		if d.Nationality != "" && !seen[d.Nationality] {
			seen[d.Nationality] = true
			out = append(out, d.Nationality)
		}
	}
	sort.Strings(out)
	return out
}

// Decades renvoie la liste des décennies de naissance présentes.
func Decades(drivers []api.Driver) []int {
	seen := map[int]bool{}
	var out []int
	for _, d := range drivers {
		if dec := d.Decade(); dec != 0 && !seen[dec] {
			seen[dec] = true
			out = append(out, dec)
		}
	}
	sort.Ints(out)
	return out
}

func contains(list []string, v string) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}
	return false
}
