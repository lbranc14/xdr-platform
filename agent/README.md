# XDR Agent - Collecteur d'événements

Agent léger de collecte d'événements pour la plateforme XDR.

## Fonctionnalités

### Collecteurs disponibles
- **System Collector** : Métriques système (CPU, mémoire, disque)
- **Network Collector** : Connexions réseau actives
- **Process Collector** : Informations sur tous les processus

### Caractéristiques
- Collecte périodique configurable
- Envoi vers Kafka en temps réel
- Heartbeat automatique
- Arrêt gracieux
- Logging détaillé

## Installation

### Prérequis
- Go 1.21+
- Accès à Kafka (localhost:9092 par défaut)

### Compilation

```bash
# Installer les dépendances
go mod download

# Compiler l'agent
go build -o xdr-agent .

# Compiler pour différentes plateformes
# Windows
GOOS=windows GOARCH=amd64 go build -o xdr-agent.exe .

# Linux
GOOS=linux GOARCH=amd64 go build -o xdr-agent-linux .

# macOS
GOOS=darwin GOARCH=amd64 go build -o xdr-agent-mac .
```

## Configuration

L'agent se configure via des variables d'environnement :

```bash
# Agent configuration
export AGENT_ID=agent-001
export AGENT_VERSION=1.0.0
export AGENT_COLLECTION_INTERVAL=30s
export AGENT_HEARTBEAT_INTERVAL=60s

# Kafka configuration
export KAFKA_BROKERS=localhost:9092
export KAFKA_TOPIC_RAW_EVENTS=raw-events

# Collectors activation
export ENABLE_SYSTEM_COLLECTOR=true
export ENABLE_NETWORK_COLLECTOR=true
export ENABLE_PROCESS_COLLECTOR=true

# Logging
export LOG_LEVEL=info
```

## Utilisation

### Lancer l'agent

```bash
# Avec les valeurs par défaut
./xdr-agent

# Avec configuration personnalisée
AGENT_ID=my-agent KAFKA_BROKERS=kafka.example.com:9092 ./xdr-agent
```

### Arrêter l'agent

Appuyez sur `Ctrl+C` pour un arrêt gracieux.

## Architecture

```
main.go
├── Charge la configuration
├── Initialise les collecteurs
├── Initialise le shipper Kafka
├── Boucle de collecte périodique
│   ├── System Collector
│   ├── Network Collector
│   └── Process Collector
└── Envoie vers Kafka
```

## Événements collectés

### Format JSON

```json
{
  "timestamp": "2024-01-02T10:30:00Z",
  "agent_id": "agent-001",
  "hostname": "server-01",
  "event_type": "process",
  "severity": "low",
  "process_name": "chrome.exe",
  "process_pid": 1234,
  "username": "user",
  "raw_data": {
    "process": {
      "pid": 1234,
      "name": "chrome.exe",
      "cpu_percent": 5.2,
      "memory_percent": 10.5
    }
  },
  "tags": ["process_monitoring", "high_cpu"]
}
```

## Développement

### Structure du code

```
agent/
├── main.go              # Point d'entrée
├── config/
│   └── config.go       # Configuration
├── models/
│   └── event.go        # Structures de données
├── collectors/
│   ├── system.go       # Collecteur système
│   ├── network.go      # Collecteur réseau
│   └── process.go      # Collecteur processus
├── shipper/
│   └── kafka.go        # Envoi Kafka
└── utils/
    └── logger.go       # Logging
```

### Ajouter un nouveau collecteur

1. Créer un fichier dans `collectors/`
2. Implémenter l'interface `Collector` avec la méthode `Collect()`
3. Retourner des `[]*models.Event`
4. L'ajouter dans `main.go`

## Tests

```bash
# Tester la compilation
go build .

# Tester avec un Kafka local
docker-compose up -d kafka
./xdr-agent

# Voir les événements dans Kafka
docker exec -it xdr-kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic raw-events \
  --from-beginning
```

## Performance

- **Empreinte mémoire** : ~50-100 MB
- **CPU** : <5% en collecte normale
- **Débit** : 1000+ événements/seconde

## Sécurité

- Pas de collecte de données sensibles (mots de passe, clés)
- Communication avec Kafka non chiffrée par défaut (ajout SSL possible)
- Logs ne contiennent pas de PII

## Dépannage

### L'agent ne démarre pas

```bash
# Vérifier la configuration
go run main.go

# Vérifier la connexion Kafka
telnet localhost 9092
```

### Pas d'événements envoyés

```bash
# Vérifier les logs
export LOG_LEVEL=debug
./xdr-agent

# Vérifier que Kafka est accessible
docker logs xdr-kafka
```

### Haute utilisation CPU/mémoire

- Augmenter `AGENT_COLLECTION_INTERVAL`
- Désactiver certains collecteurs
- Réduire le nombre de processus monitorés

## Roadmap

- [ ] Support SSL/TLS pour Kafka
- [ ] Collecteur de fichiers (File Integrity Monitoring)
- [ ] Collecteur de registre Windows
- [ ] Collecteur de logs (syslog, EventLog)
- [ ] Compression des événements
- [ ] Buffer local en cas de panne Kafka
- [ ] Métriques Prometheus

## License

MIT License
