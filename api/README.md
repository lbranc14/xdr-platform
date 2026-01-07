# XDR Ingestion Service

Service de consommation Kafka et d'insertion dans TimescaleDB.

## Fonctionnalités

- **Consumer Kafka** : Consomme les événements depuis le topic `raw-events`
- **Insertion batch** : Insère les événements par lots dans TimescaleDB
- **Multi-threaded** : Plusieurs workers en parallèle
- **Buffer intelligent** : Flush automatique par taille ou intervalle
- **Métriques** : Suivi des événements traités, insérés et erreurs
- **Arrêt gracieux** : Flush du buffer avant arrêt

## Architecture

```
Kafka (raw-events)
       ↓
Consumer (4 workers)
       ↓
Buffer (batch de 100)
       ↓
TimescaleDB (raw_events table)
```

## Installation

### Prérequis
- Go 1.21+
- Kafka accessible (localhost:9092)
- TimescaleDB accessible (localhost:5432)

### Compilation

```bash
cd api
go mod download
go build -o ingestion-service .
```

## Configuration

Via variables d'environnement :

```bash
# Database
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
export DATABASE_NAME=xdr_events
export DATABASE_USER=xdr_admin
export DATABASE_PASSWORD=xdr_secure_password_2024

# Kafka
export KAFKA_BROKERS=localhost:9092
export KAFKA_TOPIC_RAW_EVENTS=raw-events
export KAFKA_GROUP_ID=xdr-ingestion-service

# Logging
export LOG_LEVEL=info
```

## Utilisation

```bash
# Lancer le service
./ingestion-service

# Lancer avec configuration personnalisée
DATABASE_HOST=db.example.com KAFKA_BROKERS=kafka1:9092,kafka2:9092 ./ingestion-service
```

## Métriques

Le service affiche des métriques toutes les 30 secondes :

```
[INGESTION] Metrics: events_processed=1250, events_inserted=1250, errors=0, buffer_size=0
[INGESTION] Total events in database: 5432
```

## Performance

- **Débit** : ~5000 événements/seconde
- **Latence** : <100ms (batch de 100)
- **Empreinte mémoire** : ~100-200 MB
- **CPU** : ~10-20% (4 workers)

## Dépannage

### Le service ne se connecte pas à Kafka

```bash
# Vérifier que Kafka est accessible
telnet localhost 9092

# Vérifier les logs Kafka
docker logs xdr-kafka
```

### Le service ne se connecte pas à TimescaleDB

```bash
# Vérifier que TimescaleDB est accessible
psql -h localhost -p 5432 -U xdr_admin -d xdr_events

# Vérifier les logs TimescaleDB
docker logs xdr-timescaledb
```

### Les événements ne sont pas insérés

- Vérifier que l'agent envoie bien des événements vers Kafka
- Vérifier les logs du service pour les erreurs
- Augmenter le `LOG_LEVEL` à `debug`

## Développement

### Structure du code

```
api/
├── main.go              # Point d'entrée
├── config/
│   └── config.go       # Configuration
├── models/
│   └── event.go        # Structures de données
├── database/
│   └── timescale.go    # Opérations TimescaleDB
└── ingestion/
    └── consumer.go     # Consumer Kafka
```

### Tests

```bash
# Insérer des événements de test dans Kafka
cd ../agent
./xdr-agent.exe

# Vérifier que les événements sont insérés
psql -h localhost -p 5432 -U xdr_admin -d xdr_events -c "SELECT COUNT(*) FROM raw_events;"
```

## Évolutions futures

- [ ] Métriques Prometheus
- [ ] Support de plusieurs topics
- [ ] Retry logic pour les erreurs d'insertion
- [ ] Compression des événements
- [ ] Filtrage des événements en entrée
- [ ] Dead letter queue pour les événements invalides

## License

MIT License
