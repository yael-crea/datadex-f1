# 🏁 DataDex F1

> L'encyclopédie des pilotes et écuries de Formule 1, de 1950 à aujourd'hui, en Go.

Projet individuel **DataDex : fetch & show** (Ynov Campus Aix). Le site récupère ses données depuis l'API REST **Jolpica-F1**, puis fait la recherche, les filtres, la pagination et les favoris **à la main en Go**, sans aucun JavaScript.

---

## 📖 Présentation

### Contexte
Le but du projet est d'exploiter une API REST qui renvoie du JSON pour alimenter un site web. Toute la manipulation des données se fait côté serveur en Go. Les fonctions de recherche, de filtre et de pagination de l'API ne sont **pas** utilisées.

### Thème
La **Formule 1** : 885 pilotes et 214 écuries depuis 1950, plus le classement et le calendrier de la saison en cours.

### Fonctionnalités

| # | Fonctionnalité | Où ? | Détails |
|---|---|---|---|
| FT1 | **Recherche** | `/recherche`, `/pilotes`, `/ecuries` | Pilotes : prénom, nom, code (`HAM`), numéro, nationalité. Écuries : nom, nationalité. Ignore accents et majuscules, plusieurs mots possibles (« lewis ham »). |
| FT2 | **Filtres cumulables** | `/pilotes` | Nationalité + décennie de naissance + numéro permanent (avec / sans), plus un tri (nom, âge, numéro). |
| FT3 | **Pagination** | `/pilotes`, `/recherche`, `/ecuries` | 10, 20, 30 ou 50 éléments par page, au choix. Conserve les filtres et corrige les valeurs invalides. |
| FT4 | **Favoris** | `/favoris` + bouton ☆ | Ajout, retrait, vidage. Liste commune à tous, stockée dans `data/favoris.json`. |
| Bonus | Écuries, calendrier, classement | `/ecuries`, `/calendrier`, `/` | Couleurs des écuries, prochain Grand Prix, top 10 du championnat. |
| Bonus | Robustesse | partout | Cache mémoire d'une heure, copie de secours si l'API tombe, pages d'erreur 404 / 502 / 503 / 500. |

---

## ⚙️ Installation et lancement

### Prérequis
- **Go 1.22 ou plus récent** (le projet utilise les motifs de routes `GET /pilotes/{id}`) : <https://go.dev/dl/>
- Une connexion Internet (pour joindre l'API)
- Aucune dépendance externe, uniquement la bibliothèque standard

### Lancer le projet
```bash
git clone <url-du-depot> datadex-f1
cd datadex-f1
go run .
```
Ouvrir ensuite <http://localhost:8080>.

Pour changer de port :
```bash
PORT=3000 go run .          # macOS / Linux / Git Bash
$env:PORT=3000; go run .    # PowerShell
```

Pour compiler un exécutable :
```bash
go build -o datadex-f1 .
./datadex-f1
```

> ℹ️ Le serveur doit être lancé **depuis la racine du projet** : il y lit les dossiers `templates/` et `static/`, et crée le fichier `data/favoris.json` au premier démarrage.

---

## 🗂️ Structure du projet

```
datadex-f1/
├── main.go                 # Point d'entrée : configuration et démarrage du serveur
├── api/
│   ├── client.go           # Appels HTTP, cache, limitation du débit, erreurs
│   ├── f1.go               # Une fonction par endpoint de l'API
│   └── models.go           # Structures Go (Driver, Constructor, Race…)
├── collection/             # Manipulation des données « à la main »
│   ├── search.go           # FT1 : recherche
│   ├── filter.go           # FT2 : filtres et tri
│   └── pagination.go       # FT3 : pagination générique
├── favorites/
│   └── store.go            # FT4 : favoris persistants (JSON + mutex)
├── handlers/               # Routes et contrôleurs des pages
├── templates/              # Templates HTML (layout avec header et footer + pages)
├── static/css/style.css    # Feuille de style
└── data/                   # favoris.json (créé automatiquement, ignoré par Git)
```

---

## 🛣️ Routes

| Méthode | Route | Description |
|---|---|---|
| GET | `/` | Accueil : présentation, chiffres clés, prochain Grand Prix, top 10 du championnat |
| GET | `/pilotes` | Collection des pilotes avec recherche, filtres, tri et pagination |
| GET | `/pilotes/{id}` | Fiche d'un pilote : infos, départs, victoires, taux de victoire, saisons, écuries |
| GET | `/ecuries` | Collection des écuries avec recherche, filtre par nationalité et pagination |
| GET | `/calendrier` | Calendrier de la saison en cours (terminé / prochain / à venir) |
| GET | `/recherche?q=…` | Résultats de recherche (pilotes paginés + écuries) |
| GET | `/favoris` | Liste des favoris et gestion |
| POST | `/favoris/ajouter` | Ajoute un pilote (`id`, `retour`) |
| POST | `/favoris/supprimer` | Retire un pilote (`id`, `retour`) |
| POST | `/favoris/vider` | Vide la liste des favoris |
| GET | `/a-propos` | Présentation du projet et FAQ de retour d'expérience |
| GET | `/static/…` | Fichiers statiques (CSS) |
| * | toute autre URL | Page 404 personnalisée |

### Paramètres d'URL

| Paramètre | Pages | Valeurs |
|---|---|---|
| `q` | `/pilotes`, `/ecuries`, `/recherche` | texte (80 caractères maximum) |
| `nationalite` | `/pilotes`, `/ecuries` | nationalité de l'API (ex. `French`) |
| `decennie` | `/pilotes` | `1890` à `2020` (multiple de 10) |
| `numero` | `/pilotes` | `avec` ou `sans` |
| `tri` | `/pilotes` | `nom`, `recent`, `ancien`, `numero` |
| `page` | pages paginées | entier ≥ 1 (au-delà du maximum : dernière page) |
| `parpage` | pages paginées | `10`, `20`, `30`, `50` (sinon `20`) |

Une valeur invalide n'entraîne jamais d'erreur : elle est simplement ignorée.

---

## 🔌 API utilisée : Jolpica-F1

- **Documentation** : <https://github.com/jolpica/jolpica-f1>
- **URL de base** : `https://api.jolpi.ca/ergast/f1`
- **Format** : JSON, architecture REST, **aucune clé requise**
- **Quotas** : 4 requêtes par seconde, 500 par heure. Ils sont respectés grâce au cache d'une heure et à un délai de 300 ms entre deux appels.

| Endpoint | Utilisé pour | Page(s) |
|---|---|---|
| `GET /drivers.json?limit=100&offset=N` | Télécharger **tous** les pilotes (9 requêtes) | `/pilotes`, `/recherche`, `/` |
| `GET /drivers/{id}.json` | Infos d'un pilote, vérification avant l'ajout en favori | `/pilotes/{id}`, `/favoris/ajouter` |
| `GET /drivers/{id}/seasons.json` | Saisons disputées par le pilote | `/pilotes/{id}` |
| `GET /drivers/{id}/constructors.json` | Écuries du pilote | `/pilotes/{id}` |
| `GET /drivers/{id}/results.json?limit=1` | Nombre de départs (champ `total`) | `/pilotes/{id}` |
| `GET /drivers/{id}/results/1.json?limit=1` | Nombre de victoires (champ `total`) | `/pilotes/{id}` |
| `GET /constructors.json?limit=100&offset=N` | Télécharger **toutes** les écuries | `/ecuries`, `/recherche`, `/` |
| `GET /current/driverstandings.json` | Classement pilotes de la saison | `/` |
| `GET /current/races.json` | Calendrier de la saison | `/calendrier`, `/` |

> ⚠️ `limit` et `offset` servent **seulement à télécharger la collection complète**, parce que l'API renvoie 100 éléments au maximum par requête. La pagination montrée à l'utilisateur est entièrement calculée en Go (`collection/pagination.go`). Aucun paramètre de recherche ou de filtre de l'API n'est utilisé.

---

## 🛡️ Gestion des erreurs

| Situation | Comportement |
|---|---|
| API injoignable ou réponse 5xx | Copie du cache si elle existe, sinon page **502 « Drapeau rouge »** |
| API saturée (429) | Page **503 « Drapeau jaune »** |
| Pilote inexistant (`/pilotes/xyz`) | Page **404** |
| Route inexistante | Page **404** |
| Identifiant invalide envoyé au formulaire de favoris | Page **400** |
| Paramètre d'URL invalide (`?page=abc`) | Ignoré, valeur par défaut |
| Fichier `favoris.json` corrompu | Mis de côté dans `favoris.json.corrompu`, liste repartie de zéro |
| Panique dans un handler | Interceptée, page **500**, le serveur continue de tourner |
| Accueil avec une partie des données manquante | La page s'affiche quand même, avec un message d'avertissement |

---

## 👤 Auteur

Projet réalisé dans le cadre du cursus Ynov Campus Aix, 2026.
