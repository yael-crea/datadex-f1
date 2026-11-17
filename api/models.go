package api

import (
	"strconv"
	"strings"
	"time"
)

// response reproduit l'enveloppe "MRData" commune à toutes les réponses.
type response struct {
	MRData struct {
		Total            string `json:"total"`
		DriverTable      *struct{ Drivers []Driver } `json:"DriverTable"`
		ConstructorTable *struct {
			Constructors []Constructor
		} `json:"ConstructorTable"`
		SeasonTable *struct{ Seasons []Season } `json:"SeasonTable"`
		RaceTable   *struct {
			Season string
			Races  []Race
		} `json:"RaceTable"`
		StandingsTable *struct {
			Season         string `json:"season"`
			Round          string `json:"round"`
			StandingsLists []struct {
				DriverStandings []DriverStanding
			}
		} `json:"StandingsTable"`
	} `json:"MRData"`
}

func (r *response) total() int {
	n, _ := strconv.Atoi(r.MRData.Total)
	return n
}

// Driver représente un pilote de Formule 1.
type Driver struct {
	DriverID        string `json:"driverId"`
	PermanentNumber string `json:"permanentNumber"`
	Code            string `json:"code"`
	URL             string `json:"url"`
	GivenName       string `json:"givenName"`
	FamilyName      string `json:"familyName"`
	DateOfBirth     string `json:"dateOfBirth"`
	Nationality     string `json:"nationality"`
}

// FullName renvoie « Prénom Nom ».
func (d Driver) FullName() string { return d.GivenName + " " + d.FamilyName }

// BirthDate convertit la date de naissance (format AAAA-MM-JJ).
func (d Driver) BirthDate() (time.Time, bool) {
	t, err := time.Parse("2006-01-02", d.DateOfBirth)
	return t, err == nil
}

// BirthYear renvoie l'année de naissance, 0 si inconnue.
func (d Driver) BirthYear() int {
	if t, ok := d.BirthDate(); ok {
		return t.Year()
	}
	return 0
}

// Decade renvoie la décennie de naissance (ex : 1985 -> 1980).
func (d Driver) Decade() int {
	y := d.BirthYear()
	return y - y%10
}

// Initials renvoie les initiales, affichées quand le pilote n'a pas de numéro.
func (d Driver) Initials() string {
	out := ""
	for _, part := range []string{d.GivenName, d.FamilyName} {
		if r := []rune(strings.TrimSpace(part)); len(r) > 0 {
			out += string(r[0])
		}
	}
	return strings.ToUpper(out)
}

// Constructor représente une écurie.
type Constructor struct {
	ConstructorID string `json:"constructorId"`
	URL           string `json:"url"`
	Name          string `json:"name"`
	Nationality   string `json:"nationality"`
}

// Season représente une saison disputée.
type Season struct {
	Season string `json:"season"`
	URL    string `json:"url"`
}

// Race représente un Grand Prix du calendrier.
type Race struct {
	Season   string `json:"season"`
	Round    string `json:"round"`
	URL      string `json:"url"`
	RaceName string `json:"raceName"`
	Circuit  struct {
		CircuitID   string `json:"circuitId"`
		CircuitName string `json:"circuitName"`
		Location    struct {
			Locality string `json:"locality"`
			Country  string `json:"country"`
		} `json:"Location"`
	} `json:"Circuit"`
	Date string `json:"date"`
	Time string `json:"time"`
}

// Day renvoie la date de la course.
func (r Race) Day() (time.Time, bool) {
	t, err := time.Parse("2006-01-02", r.Date)
	return t, err == nil
}

// DriverStanding est une ligne du classement pilotes.
type DriverStanding struct {
	Position     string        `json:"position"`
	Points       string        `json:"points"`
	Wins         string        `json:"wins"`
	Driver       Driver        `json:"Driver"`
	Constructors []Constructor `json:"Constructors"`
}

// Team renvoie l'écurie actuelle du pilote dans le classement.
func (s DriverStanding) Team() Constructor {
	if len(s.Constructors) == 0 {
		return Constructor{}
	}
	return s.Constructors[len(s.Constructors)-1]
}
