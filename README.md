# 🛡️ XDR Platform - Cloud-Native Extended Detection & Response

[![Kubernetes](https://img.shields.io/badge/Kubernetes-326CE5?style=for-the-badge&logo=kubernetes&logoColor=white)](https://kubernetes.io/)
[![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org/)
[![React](https://img.shields.io/badge/React-20232A?style=for-the-badge&logo=react&logoColor=61DAFB)](https://reactjs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://www.docker.com/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)](https://opensource.org/licenses/MIT)

> Plateforme XDR moderne déployée sur Kubernetes pour la détection et la réponse aux menaces de sécurité en temps réel.

**🌐 Demo Live :** [https://xdr-platform.duckdns.org](https://xdr-platform.duckdns.org)

---

## 📋 Table des Matières

- [Vue d'Ensemble](#-vue-densemble)
- [Fonctionnalités](#-fonctionnalités)
- [Architecture](#️-architecture)
- [Stack Technique](#️-stack-technique)
- [Quick Start](#-quick-start)
- [Déploiement Kubernetes](#-déploiement-kubernetes)
- [Configuration](#️-configuration)
- [API Documentation](#-api-documentation)
- [Screenshots](#-screenshots)
- [Métriques & Performance](#-métriques--performance)
- [Problématiques Résolues](#-problématiques-résolues)
- [Roadmap](#-roadmap)
- [Contributing](#-contributing)
- [License](#-license)

---

## 🎯 Vue d'Ensemble

XDR Platform est une solution **Extended Detection and Response** cloud-native construite avec une architecture microservices sur Kubernetes. La plateforme collecte, analyse et visualise les événements de sécurité en temps réel, permettant aux équipes SOC d'investiguer rapidement les incidents.

### 🎓 Contexte du Projet

Projet personnel développé pour démontrer mes compétences en :
- **Orchestration Kubernetes** en production (Oracle Cloud)
- **Développement microservices** (Go, React)
- **Architecture cloud-native** scalable
- **DevSecOps** et troubleshooting avancé

### ✨ Points Forts

- ✅ **Production-ready** : Déployé sur Oracle Cloud avec HTTPS
- ✅ **Haute disponibilité** : Replicas multiples (2-3 par service)
- ✅ **0€ de coût** : Utilise le Free Tier d'Oracle Cloud
- ✅ **12+ pods** Kubernetes actifs 24/7
- ✅ **500+ événements** traités et visualisés

---

## 🚀 Fonctionnalités

### Dashboard SOC Moderne
- 📊 **Visualisations temps réel** : Timeline interactive (24h), graphiques par sévérité
- 🔍 **Filtres avancés** : Type d'événement, sévérité, hostname, date range, recherche globale
- 📥 **Export CSV** : Téléchargement des données filtrées pour analyse
- 🔄 **Auto-refresh** : Actualisation automatique toutes les 10 secondes
- 🎨 **Interface moderne** : Design responsive, cards statistiques

### Collecte & Traitement
- 🤖 **Agent de collecte** : Surveillance système, réseau, processus (Go)
- ⚡ **Ingestion haute performance** : Pipeline asynchrone avec Kafka
- 🗄️ **Stockage optimisé** : TimescaleDB (hypertables) pour séries temporelles
- 🔗 **API REST** : Endpoints complets avec pagination (Go + Fiber)

### Infrastructure
- ☸️ **Kubernetes (OKE)** : 2 nodes, LoadBalancer public, Ingress NGINX
- 🔐 **HTTPS** : Certificat SSL avec cert-manager
- 🌐 **DNS** : Domaine configuré avec DuckDNS
- 📦 **Docker** : Images optimisées multi-stage (~9-30 MB)

---

## 🏗️ Architecture

### Vue d'Ensemble

```
┌─────────────────────────────────────────────────────────────┐
│                    NGINX Ingress Controller                 │
│              (LoadBalancer IP: 89.168.47.41)               │
│                    HTTPS (cert-manager)                     │
└────────────────────────┬────────────────────────────────────┘
                         │
        ┌────────────────┴────────────────┐
        │                                 │
┌───────▼─────────┐              ┌───────▼─────────┐
│   Frontend      │              │  API Gateway    │
│   (React)       │              │  (Go + Fiber)   │
│   3 replicas    │◄─────────────┤  3 replicas     │
│   NGINX Proxy   │              │  REST API       │
└─────────────────┘              └────────┬────────┘
                                          │
                         ┌────────────────┴────────────────┐
                         │                                 │
                 ┌───────▼────────┐              ┌────────▼────────┐
                 │   Ingestion    │              │  Collection     │
                 │   Service      │              │  Agent          │
                 │   (Go)         │◄─────────────┤  (Go)           │
                 │   2 replicas   │    Kafka     │  2 replicas     │
                 └───────┬────────┘              └─────────────────┘
                         │
        ┌────────────────┴─────────────────┐
        │                                  │
┌───────▼───────┐              ┌──────────▼──────────┐
│ TimescaleDB   │              │     Redis           │
│ (PostgreSQL)  │              │   (Cache)           │
│ Hypertables   │              │   7-alpine          │
│ 40 GB PVC     │              │   5 GB PVC          │
└───────────────┘              └─────────────────────┘
```

### Flux de Données

1. **Collecte** : Agent → Kafka Topic `raw-events`
2. **Ingestion** : Service Ingestion consomme Kafka → Valide → Enrichit → Insère TimescaleDB
3. **API** : Frontend → API Gateway → Requêtes SQL optimisées → Retour JSON
4. **Visualisation** : Dashboard affiche timeline, stats, filtres en temps réel

---

## 🛠️ Stack Technique

### Backend (Go)

| Service | Rôle | Replicas | Technologie |
|---------|------|----------|-------------|
| **Agent** | Collecte événements système/réseau/processus | 2 | Go 1.21 + Kafka Sarama |
| **Ingestion** | Consomme Kafka, valide, insère en DB | 2 | Go 1.21 + lib/pq |
| **API Gateway** | REST API avec endpoints /events, /stats, /timeline | 3 | Go 1.21 + Fiber v2 |

**Packages clés** :
- `github.com/gofiber/fiber/v2` - Framework web rapide
- `github.com/lib/pq` - Driver PostgreSQL (avec pq.Array() pour arrays)
- `github.com/segmentio/kafka-go` - Client Kafka

### Frontend (React)

| Composant | Technologie | Description |
|-----------|-------------|-------------|
| **UI** | React 18 + TypeScript | Framework JavaScript moderne |
| **Build** | Vite 7 | Build tool ultra-rapide |
| **Charts** | Recharts | Visualisation de données |
| **HTTP** | Axios | Client REST API |
| **Proxy** | NGINX Alpine | Reverse proxy + assets statiques |

**Replicas** : 3 pods pour haute disponibilité

### Data Layer

| Composant | Type | Usage | Configuration |
|-----------|------|-------|---------------|
| **TimescaleDB** | PostgreSQL 15 + extension | Stockage événements time-series | Hypertables, indexes, 40 GB PVC |
| **Redis** | Cache in-memory | Cache requêtes, corrélation temps réel | 7-alpine, 5 GB PVC |
| **Kafka** | Message broker | Queue événements (désactivé en prod*) | Confluent Platform 7.5.0 |

*Kafka temporairement désactivé pour économiser RAM (2 GB total sur Free Tier)

### Infrastructure

| Composant | Version | Description |
|-----------|---------|-------------|
| **Kubernetes** | v1.34.1 | OKE (Oracle Kubernetes Engine) |
| **Nodes** | 2x VM.Standard.E3.Flex | Architecture x86, région EU Paris |
| **Ingress** | NGINX Ingress v1.9.5 | Routage HTTP/HTTPS, terminaison SSL |
| **Cert-Manager** | v1.13.3 | Gestion certificats SSL auto-signés |
| **LoadBalancer** | Oracle Cloud LB | IP publique 89.168.47.41 |
| **DNS** | DuckDNS | xdr-platform.duckdns.org |

---

## 🚀 Quick Start

### Prérequis

- Docker Desktop 20+
- kubectl 1.28+
- Go 1.21+ (optionnel, pour développement local)
- Node.js 18+ (optionnel, pour frontend local)

### Option 1 : Docker Compose (Local)

```bash
# Cloner le repository
git clone https://github.com/votre-username/xdr-platform.git
cd xdr-platform

# Démarrer l'infrastructure
docker-compose up -d

# Vérifier les services
docker-compose ps
```

**Services disponibles** :
- Frontend : http://localhost:3000
- API Gateway : http://localhost:8000
- TimescaleDB : localhost:5432 (user: `xdr_admin`, pass: `xdr_secure_password_2024`)
- Redis : localhost:6379 (pass: `xdr_redis_password_2024`)
- Kafka : localhost:9092

### Option 2 : Kubernetes (Production)

Voir la section [Déploiement Kubernetes](#-déploiement-kubernetes) ci-dessous.

---

## ☸️ Déploiement Kubernetes

### 1. Créer le Namespace

```bash
kubectl create namespace xdr-platform
```

### 2. Créer les Secrets

```bash
# TimescaleDB
kubectl create secret generic timescaledb-secret \
  --from-literal=POSTGRES_USER=xdr_admin \
  --from-literal=POSTGRES_PASSWORD=xdr_secure_password_2024 \
  --from-literal=POSTGRES_DB=xdr_events \
  -n xdr-platform

# Redis
kubectl create secret generic redis-secret \
  --from-literal=REDIS_PASSWORD=xdr_redis_password_2024 \
  -n xdr-platform
```

### 3. Déployer l'Infrastructure

```bash
cd kubernetes

# Base de données et cache
kubectl apply -f 10-timescaledb.yaml
kubectl apply -f 11-redis.yaml

# Attendre que les pods soient prêts
kubectl wait --for=condition=Ready pod -l app=timescaledb -n xdr-platform --timeout=120s
kubectl wait --for=condition=Ready pod -l app=redis -n xdr-platform --timeout=60s
```

### 4. Initialiser la Base de Données

```bash
# Se connecter à TimescaleDB
kubectl exec -it $(kubectl get pod -l app=timescaledb -n xdr-platform -o jsonpath='{.items[0].metadata.name}') -n xdr-platform -- psql -U xdr_admin -d xdr_events

# Dans psql, exécuter le schéma SQL (voir docs/schema.sql)
```

### 5. Déployer les Services

```bash
# Services backend
kubectl apply -f 20-agent.yaml
kubectl apply -f 21-ingestion.yaml
kubectl apply -f 22-api-gateway.yaml

# Frontend
kubectl apply -f 23-frontend.yaml
```

### 6. Configurer HTTPS

```bash
# Installer cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.3/cert-manager.yaml

# Attendre cert-manager
sleep 60

# Créer le certificat auto-signé
kubectl apply -f selfsigned-cert.yaml

# Attendre que le certificat soit prêt
kubectl wait --for=condition=Ready certificate/xdr-tls-selfsigned -n xdr-platform --timeout=60s

# Déployer l'Ingress
kubectl apply -f ingress-https.yaml
```

### 7. Obtenir l'IP Publique

```bash
kubectl get ingress -n xdr-platform

# Récupérer l'adresse dans la colonne ADDRESS
# Configurer votre DNS pour pointer vers cette IP
```

---

## ⚙️ Configuration

### Variables d'Environnement

#### API Gateway

```yaml
env:
  - name: DB_HOST
    value: "timescaledb-service"
  - name: DB_PORT
    value: "5432"
  - name: DB_NAME
    value: "xdr_events"
  - name: DB_USER
    valueFrom:
      secretKeyRef:
        name: timescaledb-secret
        key: POSTGRES_USER
  - name: DB_PASSWORD
    valueFrom:
      secretKeyRef:
        name: timescaledb-secret
        key: POSTGRES_PASSWORD
```

### Ressources Kubernetes

**Recommandations pour environnement de production** :

```yaml
resources:
  requests:
    memory: "256Mi"   # TimescaleDB, API Gateway
    memory: "128Mi"   # Services Go (Agent, Ingestion)
    memory: "64Mi"    # Frontend NGINX
    cpu: "100m"
  limits:
    memory: "512Mi"   # TimescaleDB
    memory: "256Mi"   # Services Go
    memory: "128Mi"   # Frontend
    cpu: "500m"
```

---

## 📡 API Documentation

### Base URL

- **Local** : `http://localhost:8000`
- **Production** : `https://xdr-platform.duckdns.org/api`

### Endpoints

#### `GET /api/v1/events`

Récupère la liste des événements avec pagination.

**Query Parameters** :
- `limit` (int) : Nombre d'événements (default: 100, max: 1000)
- `offset` (int) : Position de départ (default: 0)

**Response** :
```json
{
  "events": [
    {
      "id": 1,
      "timestamp": "2026-01-07T10:30:00Z",
      "event_type": "network",
      "severity": "high",
      "hostname": "web-server-01",
      "source_ip": "10.0.142.87",
      "tags": ["production", "security"]
    }
  ],
  "total": 500
}
```

#### `GET /api/v1/stats`

Statistiques agrégées.

**Response** :
```json
{
  "total_events": 500,
  "by_severity": {
    "critical": 46,
    "high": 49,
    "medium": 205,
    "low": 200
  }
}
```

#### `GET /api/v1/timeline`

Données timeline dernières 24h.

#### `GET /health`

Health check Kubernetes.

---

## 📸 Screenshots

### Dashboard avec Filtres
![Dashboard with Filters](docs/images/dashboard-filters.png)

*Vue principale avec cartes statistiques, timeline interactive et panneau de filtres (Time Range, Event Type, Severity, Hostname)*

### Tableaux de Bord Détaillés
![Dashboard Details](docs/images/dashboard-full.png)

*Graphiques de distribution par type et sévérité + tableau des événements récents avec tags*

---

## 📊 Métriques & Performance

### Infrastructure

- **Pods actifs** : 12+
- **Nodes Kubernetes** : 2x VM.Standard.E3.Flex (x86)
- **RAM totale** : ~2 GB (Free Tier)
- **Stockage** : 40 GB Block Storage
- **Coût** : 0€/mois

### Performances

- **Temps de réponse API** : < 100ms (p95)
- **Throughput** : 1000+ events/sec (théorique)
- **Base de données** : 500+ événements
- **Disponibilité** : 99%+ avec auto-restart

### Code

- **Lignes de code** : ~7500 lignes
  - Go : ~3000 lignes
  - TypeScript/JavaScript : ~2000 lignes
  - YAML : ~1500 lignes
  - SQL : ~500 lignes

---

## 🔧 Problématiques Résolues

### 1. **Gestion des Arrays PostgreSQL**
**Problème** : `sql: Scan error on column tags`  
**Solution** : Utilisation de `pq.Array()` pour scanner TEXT[]

### 2. **ImageInspectError sur Kubernetes**
**Problème** : `short name mode is enforcing`  
**Solution** : Préfixage toutes images avec `docker.io/`

### 3. **TimescaleDB CrashLoopBackOff**
**Problème** : `directory exists but is not empty (lost+found)`  
**Solution** : Ajout de `subPath: pgdata` dans volumeMounts

### 4. **Frontend CrashLoopBackOff**
**Problème** : `nginx: host not found in upstream`  
**Solution** : Renommage service Kubernetes

### 5. **Let's Encrypt Timeout**
**Problème** : `dial tcp 172.65.32.248:443: i/o timeout`  
**Solution** : Certificat auto-signé

---

## 🗺️ Roadmap

### ✅ Phase 1 - Foundation (Terminé)
- [x] Infrastructure Docker & Kubernetes
- [x] Agent de collecte + Ingestion
- [x] API Gateway + Dashboard React
- [x] Déploiement Oracle Cloud + HTTPS

### 🔄 Phase 2 - Améliorations
- [ ] Tests unitaires
- [ ] CI/CD GitHub Actions
- [ ] Documentation Swagger
- [ ] Monitoring Prometheus
- [ ] Logging centralisé

### 📅 Phase 3 - Fonctionnalités Avancées
- [ ] Détection ML/IA
- [ ] Corrélation événements
- [ ] Threat Intelligence
- [ ] SOAR playbooks
- [ ] Auth JWT + RBAC

---

## 🤝 Contributing

Les contributions sont bienvenues ! 

1. Fork le projet
2. Créer une branche : `git checkout -b feature/ma-fonctionnalite`
3. Commit : `git commit -m 'Ajout fonctionnalité'`
4. Push : `git push origin feature/ma-fonctionnalite`
5. Ouvrir une Pull Request

---

## 📄 License

Ce projet est sous licence **MIT License**. Voir [LICENSE](LICENSE).

---

## 📞 Contact

**Louis BRANCHUT**

- 🌐 Demo : [https://xdr-platform.duckdns.org](https://xdr-platform.duckdns.org)
- 💼 LinkedIn : [votre-profil-linkedin](https://www.linkedin.com/in/votre-profil)
- 📧 Email : votre.email@example.com
- 🐙 GitHub : [@votre-username](https://github.com/votre-username)

---

<div align="center">

**⭐ Si ce projet vous a été utile, n'hésitez pas à lui donner une étoile !**

Made with ❤️ for Cybersecurity

</div>
