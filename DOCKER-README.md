# 🐳 XDR Platform - Version Dockerisée

La plateforme XDR complète dans des containers Docker pour un déploiement simple et reproductible.

## 🚀 Quick Start

### Prérequis
- Docker Desktop installé et lancé
- 8 GB RAM minimum
- 20 GB d'espace disque

### Démarrage rapide

```powershell
# Lancer toute la stack
.\scripts\start-docker.ps1

# Ou manuellement
docker-compose up -d
```

### Accès aux services

| Service | URL | Description |
|---------|-----|-------------|
| **Dashboard** | http://localhost | Interface SOC principale |
| **API Gateway** | http://localhost:8000 | REST API |
| **Kafka UI** | http://localhost:8080 | Monitoring Kafka |
| **pgAdmin** | http://localhost:5050 | Interface PostgreSQL |

## 📁 Structure

```
xdr-platform/
├── agent/
│   ├── Dockerfile           # Image de l'agent
│   └── .dockerignore
├── api/
│   ├── Dockerfile           # Image du service d'ingestion
│   └── .dockerignore
├── api-gateway/
│   ├── Dockerfile           # Image de l'API Gateway
│   └── .dockerignore
├── frontend/
│   ├── Dockerfile           # Image du frontend
│   ├── nginx.conf          # Config Nginx
│   └── .dockerignore
├── docker-compose.yml       # Orchestration complète
└── scripts/
    └── start-docker.ps1    # Script de démarrage
```

## 🔧 Commandes Docker

### Gestion de la stack complète

```powershell
# Démarrer tous les services
docker-compose up -d

# Arrêter tous les services
docker-compose down

# Voir les logs en temps réel
docker-compose logs -f

# Voir les logs d'un service spécifique
docker-compose logs -f api-gateway

# Reconstruire les images
docker-compose build

# Reconstruire et redémarrer
docker-compose up -d --build
```

### Gestion des services individuels

```powershell
# Redémarrer un service
docker-compose restart agent

# Voir l'état des services
docker-compose ps

# Exécuter une commande dans un container
docker-compose exec api-gateway sh

# Voir les ressources utilisées
docker stats
```

### Nettoyage

```powershell
# Arrêter et supprimer tout (ATTENTION: supprime les volumes)
docker-compose down -v

# Supprimer les images inutilisées
docker image prune -a

# Nettoyer complètement Docker
docker system prune -a --volumes
```

## 🐛 Dépannage

### Les services ne démarrent pas

```powershell
# Vérifier Docker
docker info

# Voir les logs détaillés
docker-compose logs

# Vérifier les ports occupés
netstat -ano | findstr :8000
```

### Rebuild après modification du code

```powershell
# 1. Arrêter la stack
docker-compose down

# 2. Rebuild les images modifiées
docker-compose build agent api-gateway

# 3. Redémarrer
docker-compose up -d
```

### Problèmes de connexion entre services

```powershell
# Tester la connexion réseau
docker-compose exec agent ping timescaledb

# Vérifier le réseau Docker
docker network inspect xdr-platform_xdr-network
```

### Base de données vide après redémarrage

```powershell
# Les données sont dans des volumes Docker persistants
docker volume ls

# Pour repartir de zéro (ATTENTION: perte de données)
docker-compose down -v
docker-compose up -d
```

## 📊 Monitoring

### Voir les ressources utilisées

```powershell
# Stats en temps réel
docker stats

# Utilisation des volumes
docker system df -v
```

### Health checks

```powershell
# Vérifier la santé de tous les services
docker-compose ps

# Tester l'API
curl http://localhost:8000/health

# Tester le frontend
curl http://localhost/health
```

## 🔐 Sécurité

### Changement des mots de passe

Modifiez les variables d'environnement dans `docker-compose.yml` :

```yaml
environment:
  POSTGRES_PASSWORD: votre_nouveau_mot_de_passe
  REDIS_PASSWORD: votre_nouveau_mot_de_passe
```

Puis recréez les containers :

```powershell
docker-compose down -v
docker-compose up -d
```

## 🚀 Avantages de cette approche

✅ **Un seul fichier** pour tout déployer  
✅ **Isolation complète** des services  
✅ **Reproductible** sur n'importe quelle machine  
✅ **Prêt pour Kubernetes** (même architecture)  
✅ **Volumes persistants** (données conservées)  
✅ **Health checks** automatiques  
✅ **Réseaux isolés** pour la sécurité  

## 📈 Prochaines étapes

- [ ] Ajouter Prometheus pour le monitoring
- [ ] Ajouter Grafana pour la visualisation
- [ ] Configurer des backups automatiques
- [ ] Déployer sur Kubernetes
- [ ] Ajouter le support HTTPS

## 🆘 Support

En cas de problème :
1. Vérifiez les logs : `docker-compose logs`
2. Vérifiez l'état : `docker-compose ps`
3. Redémarrez : `docker-compose restart`

## 📝 License

MIT License
