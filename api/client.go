// Package api regroupe tout ce qui concerne l'appel à l'API Jolpica-F1
// (successeur de l'API Ergast) : requêtes HTTP, cache mémoire et erreurs.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// BaseURL est l'adresse racine de l'API Jolpica-F1.
const BaseURL = "https://api.jolpi.ca/ergast/f1"

// Erreurs « métier » renvoyées par le client. Les handlers s'en servent
// pour choisir le bon code HTTP et la bonne page d'erreur.
var (
	ErrNotFound    = errors.New("ressource introuvable")
	ErrUnavailable = errors.New("l'API F1 est indisponible")
	ErrRateLimited = errors.New("trop de requêtes envoyées à l'API F1")
)

// Client appelle l'API et garde les réponses en cache pour respecter
// les quotas de Jolpica (4 requêtes/seconde, 500 requêtes/heure).
type Client struct {
	http    *http.Client
	baseURL string
	ttl     time.Duration

	mu        sync.Mutex
	cache     map[string]cacheEntry
	pathLocks map[string]*sync.Mutex

	rateMu   sync.Mutex
	lastCall time.Time
	minDelay time.Duration
}

type cacheEntry struct {
	body    []byte
	expires time.Time
}

// NewClient crée un client prêt à l'emploi. La variable d'environnement
// F1_API_URL permet de changer l'adresse de l'API (utile pour tester le
// comportement du site quand l'API est hors ligne).
func NewClient() *Client {
	base := BaseURL
	if custom := os.Getenv("F1_API_URL"); custom != "" {
		base = strings.TrimRight(custom, "/")
	}
	return &Client{
		http:      &http.Client{Timeout: 10 * time.Second},
		baseURL:   base,
		ttl:       time.Hour,
		cache:     make(map[string]cacheEntry),
		pathLocks: make(map[string]*sync.Mutex),
		minDelay:  400 * time.Millisecond,
	}
}

// get renvoie le corps JSON brut de l'URL demandée.
// Ordre de priorité : cache valide > appel réseau > cache périmé (secours).
func (c *Client) get(path string) ([]byte, error) {
	// Un verrou par URL : si deux visiteurs demandent la même ressource en
	// même temps, un seul appel part vers l'API, l'autre attend le cache.
	c.mu.Lock()
	lock, ok := c.pathLocks[path]
	if !ok {
		lock = &sync.Mutex{}
		c.pathLocks[path] = lock
	}
	c.mu.Unlock()
	lock.Lock()
	defer lock.Unlock()

	c.mu.Lock()
	entry, inCache := c.cache[path]
	c.mu.Unlock()

	if inCache && time.Now().Before(entry.expires) {
		return entry.body, nil
	}

	body, err := c.fetchRemote(path)
	if err != nil {
		// Si l'API tombe mais qu'on a une ancienne copie, on l'utilise :
		// le site reste consultable même hors ligne.
		if inCache && !errors.Is(err, ErrNotFound) {
			log.Printf("api: %s indisponible, utilisation du cache périmé (%v)", path, err)
			return entry.body, nil
		}
		return nil, err
	}

	c.mu.Lock()
	c.cache[path] = cacheEntry{body: body, expires: time.Now().Add(c.ttl)}
	c.mu.Unlock()
	return body, nil
}

// maxRetries : nombre de nouvelles tentatives si l'API répond 429.
const maxRetries = 3

// fetchRemote appelle l'API ; en cas de code 429 (trop de requêtes),
// on patiente puis on réessaie au lieu d'afficher tout de suite une erreur.
func (c *Client) fetchRemote(path string) ([]byte, error) {
	for attempt := 1; ; attempt++ {
		body, err := c.doRequest(path)
		if !errors.Is(err, ErrRateLimited) || attempt > maxRetries {
			return body, err
		}
		wait := time.Duration(attempt) * 2 * time.Second // 2 s, 4 s, 6 s
		log.Printf("api: 429 sur %s, nouvelle tentative dans %s", path, wait)
		time.Sleep(wait)
	}
}

// doRequest fait réellement l'appel HTTP, en espaçant les requêtes.
func (c *Client) doRequest(path string) ([]byte, error) {
	c.rateMu.Lock()
	if wait := c.minDelay - time.Since(c.lastCall); wait > 0 {
		time.Sleep(wait)
	}
	c.lastCall = time.Now()
	c.rateMu.Unlock()

	resp, err := c.http.Get(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("%w : %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return nil, ErrNotFound
	case resp.StatusCode == http.StatusTooManyRequests:
		return nil, ErrRateLimited
	case resp.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("%w : code HTTP %d", ErrUnavailable, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w : lecture impossible (%v)", ErrUnavailable, err)
	}
	return body, nil
}

// fetch appelle l'API puis décode le JSON dans la structure envoyée.
func (c *Client) fetch(path string) (*response, error) {
	body, err := c.get(path)
	if err != nil {
		return nil, err
	}
	var r response
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("%w : JSON invalide (%v)", ErrUnavailable, err)
	}
	return &r, nil
}
