package handlers

import (
	"html/template"
	"strings"
	"time"
)

// nationalityInfo associe une nationalité (en anglais, telle que renvoyée
// par l'API) à son code pays ISO et à son libellé en français.
type nationalityInfo struct {
	ISO string
	FR  string
}

var nationalities = map[string]nationalityInfo{
	"American":          {"US", "Américaine"},
	"American-Italian":  {"US", "Italo-américaine"},
	"Argentine":         {"AR", "Argentine"},
	"Argentinian":       {"AR", "Argentine"},
	"Argentine-Italian": {"AR", "Italo-argentine"},
	"Australian":        {"AU", "Australienne"},
	"Austrian":          {"AT", "Autrichienne"},
	"Belgian":           {"BE", "Belge"},
	"Brazilian":         {"BR", "Brésilienne"},
	"British":           {"GB", "Britannique"},
	"Bulgarian":         {"BG", "Bulgare"},
	"Canadian":          {"CA", "Canadienne"},
	"Chilean":           {"CL", "Chilienne"},
	"Chinese":           {"CN", "Chinoise"},
	"Colombian":         {"CO", "Colombienne"},
	"Czech":             {"CZ", "Tchèque"},
	"Danish":            {"DK", "Danoise"},
	"Dutch":             {"NL", "Néerlandaise"},
	"East German":       {"DE", "Est-allemande"},
	"Estonian":          {"EE", "Estonienne"},
	"Finnish":           {"FI", "Finlandaise"},
	"French":            {"FR", "Française"},
	"German":            {"DE", "Allemande"},
	"Hong Kong":         {"HK", "Hongkongaise"},
	"Hungarian":         {"HU", "Hongroise"},
	"Indian":            {"IN", "Indienne"},
	"Indonesian":        {"ID", "Indonésienne"},
	"Irish":             {"IE", "Irlandaise"},
	"Italian":           {"IT", "Italienne"},
	"Japanese":          {"JP", "Japonaise"},
	"Liechtensteiner":   {"LI", "Liechtensteinoise"},
	"Malaysian":         {"MY", "Malaisienne"},
	"Mexican":           {"MX", "Mexicaine"},
	"Monegasque":        {"MC", "Monégasque"},
	"New Zealander":     {"NZ", "Néo-zélandaise"},
	"Polish":            {"PL", "Polonaise"},
	"Portuguese":        {"PT", "Portugaise"},
	"Rhodesian":         {"ZW", "Rhodésienne"},
	"Russian":           {"RU", "Russe"},
	"South African":     {"ZA", "Sud-africaine"},
	"Spanish":           {"ES", "Espagnole"},
	"Swedish":           {"SE", "Suédoise"},
	"Swiss":             {"CH", "Suisse"},
	"Thai":              {"TH", "Thaïlandaise"},
	"Uruguayan":         {"UY", "Uruguayenne"},
	"Venezuelan":        {"VE", "Vénézuélienne"},
}

// flag renvoie une petite image du drapeau (les emojis drapeaux ne
// s'affichent pas sous Windows). Images fournies par flagcdn.com.
func flag(nationality string) template.HTML {
	info, ok := nationalities[nationality]
	if !ok || len(info.ISO) != 2 {
		return "🏁"
	}
	iso := strings.ToLower(info.ISO)
	return template.HTML(`<img class="flag" src="https://flagcdn.com/w40/` + iso +
		`.png" width="20" height="15" alt="" loading="lazy">`)
}

// natFR traduit une nationalité en français (repli : valeur d'origine).
func natFR(nationality string) string {
	if info, ok := nationalities[nationality]; ok {
		return info.FR
	}
	return nationality
}

// teamColors donne la couleur officielle des écuries récentes.
var teamColors = map[string]string{
	"red_bull":     "#3671C6",
	"mercedes":     "#27F4D2",
	"ferrari":      "#E8002D",
	"mclaren":      "#FF8000",
	"aston_martin": "#229971",
	"alpine":       "#0093CC",
	"williams":     "#64C4FF",
	"rb":           "#6692FF",
	"sauber":       "#52E252",
	"audi":         "#BB0A30",
	"haas":         "#B6BABD",
	"cadillac":     "#C9B37E",
	"alphatauri":   "#5E8FAA",
	"alfa":         "#C92D4B",
	"renault":      "#FFF500",
	"racing_point": "#F596C8",
	"force_india":  "#FF80C7",
	"toro_rosso":   "#469BFF",
	"lotus_f1":     "#FFB800",
}

func teamColor(id string) string {
	if c, ok := teamColors[id]; ok {
		return c
	}
	return "#8a8f98"
}

var moisFR = []string{"janvier", "février", "mars", "avril", "mai", "juin",
	"juillet", "août", "septembre", "octobre", "novembre", "décembre"}

// dateFR convertit "1985-01-07" en "7 janvier 1985".
func dateFR(s string) string {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return s
	}
	return formatFR(t)
}

func formatFR(t time.Time) string {
	day := t.Format("2")
	if day == "1" {
		day = "1er"
	}
	return day + " " + moisFR[t.Month()-1] + " " + t.Format("2006")
}

// age calcule l'âge actuel à partir d'une date AAAA-MM-JJ.
func age(s string) int {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return 0
	}
	now := time.Now()
	a := now.Year() - t.Year()
	if now.YearDay() < t.YearDay() {
		a--
	}
	return a
}

var templateFuncs = template.FuncMap{
	"flag":      flag,
	"natFR":     natFR,
	"teamColor": teamColor,
	"dateFR":    dateFR,
	"age":       age,
	"add":       func(a, b int) int { return a + b },
	"safeCSS":   func(s string) template.CSS { return template.CSS(s) },
}
