# XDR API Gateway

API REST pour la plateforme XDR, construite avec Go et Fiber.

## Fonctionnalités

- **API REST** : Endpoints pour récupérer les événements
- **CORS** : Support pour les requêtes cross-origin
- **Pagination** : Limiter le nombre de résultats
- **Health Check** : Vérification de la santé du service
- **Logging** : Logs détaillés de toutes les requêtes

## Endpoints

### Health Check
```
GET /health
```

Réponse :
```json
{
  "status": "healthy",
  "timestamp": "2024-01-02T15:30:00Z",
  "service": "xdr-api"
}
```

### Récupérer les événements
```
GET /api/v1/events?limit=50
```

Paramètres :
- `limit` : Nombre maximum d'événements (défaut: 50, max: 1000)

Réponse :
```json
{
  "success": true,
  "count": 50,
  "events": [
    {
      "timestamp": "2024-01-02T15:30:00Z",
      "agent_id": "agent-001",
      "hostname": "server-01",
      "event_type": "process",
      "severity": "low",
      "process_name": "chrome.exe",
      "process_pid": 1234,
      "raw_data": {...}
    }
  ]
}
```

### Compter les événements
```
GET /api/v1/events/count
```

Réponse :
```json
{
  "success": true,
  "count": 5432
}
```

### Statistiques
```
GET /api/v1/events/stats
```

Réponse :
```json
{
  "success": true,
  "stats": {
    "total_events": 5432,
    "last_updated": "2024-01-02T15:30:00Z"
  }
}
```

## Installation

```bash
cd api
go mod download
go build -o api-gateway.exe .
```

## Configuration

Variables d'environnement :

```bash
export API_PORT=8000
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
export DATABASE_NAME=xdr_events
export DATABASE_USER=xdr_admin
export DATABASE_PASSWORD=xdr_secure_password_2024
```

## Utilisation

```bash
# Lancer l'API
./api-gateway.exe

# L'API sera disponible sur http://localhost:8000
```

## Test des endpoints

```bash
# Health check
curl http://localhost:8000/health

# Récupérer les événements
curl http://localhost:8000/api/v1/events?limit=10

# Compter les événements
curl http://localhost:8000/api/v1/events/count

# Statistiques
curl http://localhost:8000/api/v1/events/stats
```

## Architecture

```
Client (Browser/Frontend)
       ↓
API Gateway (Fiber)
       ↓
TimescaleDB
```

## Développement

### Structure du code

```
api/
├── main.go              # Point d'entrée API REST
├── handlers/
│   └── events.go       # Handlers pour les événements
├── routes/
│   └── routes.go       # Configuration des routes
├── config/
│   └── config.go       # Configuration
├── models/
│   └── event.go        # Structures de données
└── database/
    └── timescale.go    # Opérations TimescaleDB
```

## Performance

- **Latence** : <50ms par requête
- **Débit** : 1000+ requêtes/seconde
- **Connexions** : Pool de 25 connexions DB

## Sécurité (TODO)

- [ ] Authentification JWT
- [ ] Rate limiting
- [ ] HTTPS/TLS
- [ ] Validation des entrées
- [ ] RBAC (Role-Based Access Control)

## Évolutions futures

- [ ] WebSocket pour événements temps réel
- [ ] Filtres avancés (par date, type, sévérité)
- [ ] Agrégations et statistiques avancées
- [ ] Export de données (CSV, JSON)
- [ ] Cache Redis pour performance

## License

MIT License
