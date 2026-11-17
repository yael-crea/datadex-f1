// Package favorites gère la liste de pilotes favoris, commune à tous les
// visiteurs et sauvegardée dans un fichier JSON côté serveur.
package favorites

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Favorite est un pilote ajouté aux favoris. On garde une copie des infos
// principales pour pouvoir afficher la liste même si l'API est hors ligne.
type Favorite struct {
	ID          string    `json:"id"`
	GivenName   string    `json:"givenName"`
	FamilyName  string    `json:"familyName"`
	Code        string    `json:"code,omitempty"`
	Number      string    `json:"number,omitempty"`
	Nationality string    `json:"nationality"`
	DateOfBirth string    `json:"dateOfBirth"`
	AddedAt     time.Time `json:"addedAt"`
}

// ErrAlreadyExists est renvoyée si le pilote est déjà dans les favoris.
var ErrAlreadyExists = errors.New("ce pilote est déjà dans vos favoris")

// ErrNotFound est renvoyée si on supprime un pilote absent de la liste.
var ErrNotFound = errors.New("ce pilote n'est pas dans vos favoris")

// Store garde les favoris en mémoire et les écrit dans le fichier à chaque
// changement. Le mutex évite que deux requêtes simultanées corrompent le fichier.
type Store struct {
	path  string
	mu    sync.Mutex
	items []Favorite
}

// NewStore charge le fichier existant ou le crée s'il n'existe pas.
func NewStore(path string) (*Store, error) {
	s := &Store{path: path}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("création du dossier des favoris : %w", err)
	}

	content, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return s, s.save()
	case err != nil:
		return nil, fmt.Errorf("lecture des favoris : %w", err)
	}

	if len(content) > 0 {
		if err := json.Unmarshal(content, &s.items); err != nil {
			// Fichier corrompu : on le met de côté au lieu de faire planter le site.
			backup := path + ".corrompu"
			log.Printf("favoris : fichier illisible (%v), copie de secours dans %s", err, backup)
			_ = os.WriteFile(backup, content, 0o644)
			s.items = nil
			return s, s.save()
		}
	}
	return s, nil
}

// All renvoie une copie de la liste (du plus récent au plus ancien).
func (s *Store) All() []Favorite {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Favorite, len(s.items))
	for i, f := range s.items {
		out[len(s.items)-1-i] = f
	}
	return out
}

// IDs renvoie l'ensemble des identifiants favoris (pour afficher ★ / ☆).
func (s *Store) IDs() map[string]bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make(map[string]bool, len(s.items))
	for _, f := range s.items {
		ids[f.ID] = true
	}
	return ids
}

// Add ajoute un pilote et sauvegarde le fichier.
func (s *Store) Add(f Favorite) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.items {
		if existing.ID == f.ID {
			return ErrAlreadyExists
		}
	}
	f.AddedAt = time.Now()
	s.items = append(s.items, f)
	if err := s.save(); err != nil {
		s.items = s.items[:len(s.items)-1] // on annule en cas d'échec
		return err
	}
	return nil
}

// Remove supprime un pilote et sauvegarde le fichier.
func (s *Store) Remove(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, f := range s.items {
		if f.ID == id {
			previous := s.items
			s.items = append(append([]Favorite{}, s.items[:i]...), s.items[i+1:]...)
			if err := s.save(); err != nil {
				s.items = previous
				return err
			}
			return nil
		}
	}
	return ErrNotFound
}

// Clear vide entièrement la liste.
func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	previous := s.items
	s.items = nil
	if err := s.save(); err != nil {
		s.items = previous
		return err
	}
	return nil
}

// save écrit d'abord dans un fichier temporaire puis le renomme :
// si le serveur s'arrête pendant l'écriture, l'ancien fichier reste intact.
// À appeler avec le mutex verrouillé.
func (s *Store) save() error {
	items := s.items
	if items == nil {
		items = []Favorite{}
	}
	content, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, content, 0o644); err != nil {
		return fmt.Errorf("écriture des favoris : %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("écriture des favoris : %w", err)
	}
	return nil
}
