package api

import (
	"fmt"
	"regexp"
)

// pageSize est le maximum autorisé par l'API par requête. On s'en sert
// uniquement pour TÉLÉCHARGER toute la collection : la pagination affichée
// à l'utilisateur est faite à la main côté Go (package collection).
const pageSize = 100

var validID = regexp.MustCompile(`^[a-z0-9_]+$`)

// ValidID vérifie qu'un identifiant ne contient que des caractères sûrs
// avant de l'insérer dans une URL de l'API.
func ValidID(id string) bool { return validID.MatchString(id) }

// AllDrivers récupère les ~880 pilotes de l'histoire de la F1.
// Endpoint : GET /drivers.json
func (c *Client) AllDrivers() ([]Driver, error) {
	var all []Driver
	for offset := 0; ; offset += pageSize {
		r, err := c.fetch(fmt.Sprintf("/drivers.json?limit=%d&offset=%d", pageSize, offset))
		if err != nil {
			return nil, err
		}
		if r.MRData.DriverTable == nil {
			return nil, fmt.Errorf("%w : réponse inattendue", ErrUnavailable)
		}
		all = append(all, r.MRData.DriverTable.Drivers...)
		if offset+pageSize >= r.total() {
			return all, nil
		}
	}
}

// Driver récupère un pilote précis.
// Endpoint : GET /drivers/{driverId}.json
func (c *Client) Driver(id string) (Driver, error) {
	if !ValidID(id) {
		return Driver{}, ErrNotFound
	}
	r, err := c.fetch("/drivers/" + id + ".json")
	if err != nil {
		return Driver{}, err
	}
	if r.MRData.DriverTable == nil || len(r.MRData.DriverTable.Drivers) == 0 {
		return Driver{}, ErrNotFound
	}
	return r.MRData.DriverTable.Drivers[0], nil
}

// DriverConstructors liste les écuries pour lesquelles le pilote a couru.
// Endpoint : GET /drivers/{driverId}/constructors.json
func (c *Client) DriverConstructors(id string) ([]Constructor, error) {
	r, err := c.fetch(fmt.Sprintf("/drivers/%s/constructors.json?limit=%d", id, pageSize))
	if err != nil {
		return nil, err
	}
	if r.MRData.ConstructorTable == nil {
		return nil, nil
	}
	return r.MRData.ConstructorTable.Constructors, nil
}

// DriverSeasons liste les saisons disputées par le pilote.
// Endpoint : GET /drivers/{driverId}/seasons.json
func (c *Client) DriverSeasons(id string) ([]Season, error) {
	r, err := c.fetch(fmt.Sprintf("/drivers/%s/seasons.json?limit=%d", id, pageSize))
	if err != nil {
		return nil, err
	}
	if r.MRData.SeasonTable == nil {
		return nil, nil
	}
	return r.MRData.SeasonTable.Seasons, nil
}

// DriverRaceCount renvoie le nombre de départs en Grand Prix.
// Endpoint : GET /drivers/{driverId}/results.json (on lit le champ "total")
func (c *Client) DriverRaceCount(id string) (int, error) {
	r, err := c.fetch("/drivers/" + id + "/results.json?limit=1")
	if err != nil {
		return 0, err
	}
	return r.total(), nil
}

// DriverWins renvoie le nombre de victoires.
// Endpoint : GET /drivers/{driverId}/results/1.json (on lit le champ "total")
func (c *Client) DriverWins(id string) (int, error) {
	r, err := c.fetch("/drivers/" + id + "/results/1.json?limit=1")
	if err != nil {
		return 0, err
	}
	return r.total(), nil
}

// AllConstructors récupère les ~210 écuries de l'histoire.
// Endpoint : GET /constructors.json
func (c *Client) AllConstructors() ([]Constructor, error) {
	var all []Constructor
	for offset := 0; ; offset += pageSize {
		r, err := c.fetch(fmt.Sprintf("/constructors.json?limit=%d&offset=%d", pageSize, offset))
		if err != nil {
			return nil, err
		}
		if r.MRData.ConstructorTable == nil {
			return nil, fmt.Errorf("%w : réponse inattendue", ErrUnavailable)
		}
		all = append(all, r.MRData.ConstructorTable.Constructors...)
		if offset+pageSize >= r.total() {
			return all, nil
		}
	}
}

// CurrentStandings renvoie le classement pilotes de la saison en cours.
// Endpoint : GET /current/driverstandings.json
func (c *Client) CurrentStandings() (season, round string, standings []DriverStanding, err error) {
	r, err := c.fetch("/current/driverstandings.json?limit=" + fmt.Sprint(pageSize))
	if err != nil {
		return "", "", nil, err
	}
	t := r.MRData.StandingsTable
	if t == nil || len(t.StandingsLists) == 0 {
		return "", "", nil, nil
	}
	return t.Season, t.Round, t.StandingsLists[0].DriverStandings, nil
}

// CurrentRaces renvoie le calendrier de la saison en cours.
// Endpoint : GET /current/races.json
func (c *Client) CurrentRaces() (string, []Race, error) {
	r, err := c.fetch("/current/races.json?limit=" + fmt.Sprint(pageSize))
	if err != nil {
		return "", nil, err
	}
	if r.MRData.RaceTable == nil {
		return "", nil, nil
	}
	return r.MRData.RaceTable.Season, r.MRData.RaceTable.Races, nil
}

// Preload remplit le cache au démarrage pour que la première visite soit rapide.
func (c *Client) Preload() {
	if _, err := c.AllDrivers(); err != nil {
		fmt.Println("préchargement des pilotes impossible :", err)
	}
	if _, err := c.AllConstructors(); err != nil {
		fmt.Println("préchargement des écuries impossible :", err)
	}
	_, _, _, _ = c.CurrentStandings()
	_, _, _ = c.CurrentRaces()
}
