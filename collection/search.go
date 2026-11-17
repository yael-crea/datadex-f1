// Package collection contient la manipulation « à la main » des données :
// recherche, filtres et pagination. Aucune de ces opérations n'utilise
// les paramètres de recherche/filtre/pagination de l'API.
package collection

import (
	"strings"

	"datadex-f1/api"
)

// accents remplace les lettres accentuées pour que « perez » trouve « Pérez ».
var accents = strings.NewReplacer(
	"á", "a", "à", "a", "â", "a", "ä", "a", "ã", "a", "å", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i",
	"ó", "o", "ò", "o", "ô", "o", "ö", "o", "õ", "o", "ø", "o",
	"ú", "u", "ù", "u", "û", "u", "ü", "u",
	"ç", "c", "ñ", "n", "ý", "y", "ÿ", "y", "ß", "ss",
	"-", " ", "'", " ", ".", " ",
)

// Normalize met un texte en minuscules, sans accents ni ponctuation.
func Normalize(s string) string {
	return strings.TrimSpace(accents.Replace(strings.ToLower(s)))
}

// matches vérifie que CHAQUE mot recherché apparaît dans au moins un champ.
// Ex : « lewis ham » -> « lewis » est dans le prénom, « ham » dans le nom.
func matches(words []string, fields ...string) bool {
	haystack := Normalize(strings.Join(fields, " "))
	for _, w := range words {
		if !strings.Contains(haystack, w) {
			return false
		}
	}
	return true
}

// SearchDrivers parcourt la collection et garde les pilotes dont le prénom,
// le nom, le code (HAM, VER…), le numéro ou la nationalité correspondent.
func SearchDrivers(drivers []api.Driver, query string) []api.Driver {
	words := strings.Fields(Normalize(query))
	if len(words) == 0 {
		return drivers
	}
	var out []api.Driver
	for _, d := range drivers {
		if matches(words, d.GivenName, d.FamilyName, d.Code, d.PermanentNumber, d.Nationality) {
			out = append(out, d)
		}
	}
	return out
}

// SearchConstructors cherche une écurie par son nom ou sa nationalité.
func SearchConstructors(teams []api.Constructor, query string) []api.Constructor {
	words := strings.Fields(Normalize(query))
	if len(words) == 0 {
		return teams
	}
	var out []api.Constructor
	for _, t := range teams {
		if matches(words, t.Name, t.Nationality) {
			out = append(out, t)
		}
	}
	return out
}
