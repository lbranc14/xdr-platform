# 🔧 Guide de Troubleshooting - XDR Platform

Ce document détaille les problématiques rencontrées durant le développement et leurs solutions.

---

## 1. Gestion des Arrays PostgreSQL (Tags)

### 🔴 Problème

```
sql: Scan error on column index 11, name "tags": unsupported Scan, 
storing driver.Value type []uint8 into type *[]string
```

L'API Gateway retournait une erreur lors de la lecture des événements. Les tags étaient retournés vides `[]` au lieu de contenir les valeurs.

### 🔍 Analyse

PostgreSQL stocke les arrays avec un format spécifique `{val1,val2,val3}`. Le driver Go `lib/pq` nécessite l'utilisation de `pq.Array()` pour convertir correctement entre les types PostgreSQL et Go.

Le code utilisait :
```go
var tagsRaw interface{}
err := rows.Scan(..., &tagsRaw, ...)
event.Tags = []string{} // Workaround temporaire
```

### ✅ Solution

Remplacement dans 3 fonctions (`GetRecentEvents`, `GetFilteredEvents`, `scanEvents`) :

```go
// Import requis
import "github.com/lib/pq"

// Avant (incorrect)
var tagsRaw interface{}
err := rows.Scan(..., &tagsRaw, ...)

// Après (correct)
err := rows.Scan(..., pq.Array(&event.Tags), ...)
```

**Fichiers modifiés** :
- `api-gateway/database/timescale.go` (lignes ~177, ~416)

**Commit** :
```bash
git add api-gateway/database/timescale.go
git commit -m "fix: use pq.Array() for PostgreSQL TEXT[] tags scanning"
```

---

## 2. ImageInspectError sur Tous les Pods Kubernetes

### 🔴 Problème

```
ImageInspectError: short name mode is enforcing, but image name redis:7-alpine 
returns ambiguous list
```

Tous les pods (infrastructure + applicatifs) restaient bloqués en `ImagePullBackOff`. Aucun service ne pouvait démarrer.

### 🔍 Analyse

Kubernetes sur Oracle Cloud nécessite des **noms d'images complets** avec le registry explicite. Les noms courts (`redis:7-alpine`) sont ambigus car ils peuvent pointer vers plusieurs registries :
- docker.io/redis:7-alpine (Docker Hub)
- quay.io/redis:7-alpine (Quay.io)
- etc.

Le cluster est configuré en "short name mode enforcing" pour des raisons de sécurité.

### ✅ Solution

Préfixage de **TOUTES** les images avec `docker.io/` :

**Infrastructure** :
- `redis:7-alpine` → `docker.io/redis:7-alpine`
- `timescale/timescaledb:latest-pg15` → `docker.io/timescale/timescaledb:latest-pg15`
- `confluentinc/cp-kafka:7.5.0` → `docker.io/confluentinc/cp-kafka:7.5.0`
- `confluentinc/cp-zookeeper:7.5.0` → `docker.io/confluentinc/cp-zookeeper:7.5.0`

**Applications** :
- `lbranc14/xdr-agent:latest` → `docker.io/lbranc14/xdr-agent:latest`
- `lbranc14/xdr-ingestion:latest` → `docker.io/lbranc14/xdr-ingestion:latest`
- `lbranc14/xdr-api-gateway:latest` → `docker.io/lbranc14/xdr-api-gateway:latest`
- `lbranc14/xdr-frontend:latest` → `docker.io/lbranc14/xdr-frontend:latest`

**Fichiers modifiés** : 13 manifests YAML (tous les deployments et jobs)

**Commit** :
```bash
git add kubernetes/*.yaml
git commit -m "fix(k8s): add docker.io prefix to all images for Oracle Cloud"
```

---

## 3. TimescaleDB CrashLoopBackOff (lost+found)

### 🔴 Problème

```
initdb: error: directory "/var/lib/postgresql/data" exists but is not empty
initdb: detail: It contains a lost+found directory, perhaps due to it being a mount point.
initdb: hint: Using a mount point directly as the data directory is not recommended.
```

TimescaleDB entrait en `CrashLoopBackOff`. La base de données ne pouvait pas démarrer.

### 🔍 Analyse

Les volumes persistants (Block Storage) sur Oracle Cloud sont formatés avec un système de fichiers qui crée automatiquement un dossier `lost+found` à la racine. 

PostgreSQL refuse de s'initialiser dans un dossier non vide. Le volume était monté directement sur `/var/lib/postgresql/data`, ce qui n'est pas recommandé.

### ✅ Solution

Ajout d'un `subPath` dans le manifest Kubernetes :

```yaml
# Avant
volumeMounts:
- name: timescaledb-storage
  mountPath: /var/lib/postgresql/data

# Après
volumeMounts:
- name: timescaledb-storage
  mountPath: /var/lib/postgresql/data
  subPath: pgdata  # ← Ajout

env:
- name: PGDATA
  value: /var/lib/postgresql/data/pgdata  # ← Ajout
```

**Actions supplémentaires** :
1. Suppression du namespace complet pour nettoyer les PVCs
2. Redéploiement avec la nouvelle configuration

**Fichiers modifiés** :
- `kubernetes/10-timescaledb.yaml`

**Commit** :
```bash
git add kubernetes/10-timescaledb.yaml
git commit -m "fix(k8s): add subPath for TimescaleDB volume to avoid lost+found conflict"
```

---

## 4. Frontend CrashLoopBackOff (Résolution DNS)

### 🔴 Problème

```
nginx: [emerg] host not found in upstream "api-gateway" in 
/etc/nginx/conf.d/default.conf:26
```

Le frontend NGINX crashait au démarrage. Dashboard inaccessible.

### 🔍 Analyse

Le fichier `nginx.conf` du frontend contenait :
```nginx
proxy_pass http://api-gateway:8000;
```

Mais le service Kubernetes était nommé `api-gateway-service`. NGINX ne pouvait pas résoudre le nom DNS `api-gateway` dans le cluster.

### ✅ Solution

**Option choisie** : Renommer le service Kubernetes pour correspondre au nom dans nginx.conf

```yaml
# Dans 22-api-gateway.yaml
apiVersion: v1
kind: Service
metadata:
  name: api-gateway  # Changé de "api-gateway-service"
  namespace: xdr-platform
```

**Alternative non retenue** : Modifier nginx.conf et rebuild l'image Docker (plus long)

**Fichiers modifiés** :
- `kubernetes/22-api-gateway.yaml`

**Commit** :
```bash
git add kubernetes/22-api-gateway.yaml
git commit -m "fix(k8s): rename api-gateway service for nginx DNS resolution"
```

---

## 5. Erreur "failed to unmarshal raw_data"

### 🔴 Problème

```json
{
  "details": "failed to unmarshal raw_data: invalid character 'E' looking for beginning of value",
  "error": "Failed to retrieve events"
}
```

L'API retournait une erreur HTTP 500. Dashboard affichait "Failed to fetch events".

### 🔍 Analyse

Le script SQL de génération de données de test insérait `raw_data` comme du **texte brut** :
```sql
raw_data := 'Event 1: system on web-server-01'
```

Mais l'API Gateway s'attendait à du **JSON** et tentait de parser avec `json.Unmarshal()`.

### ✅ Solution

Réécriture du script SQL pour générer du JSON valide :

```sql
raw_data_json := jsonb_build_object(
    'event_id', i,
    'message', 'Event ' || i || ': ' || rand_event_type || ' on ' || rand_hostname,
    'details', jsonb_build_object(
        'action', rand_event_type,
        'target', rand_hostname,
        'user', rand_username
    )
);

INSERT INTO raw_events (..., raw_data) VALUES (..., raw_data_json);
```

**Actions** :
1. `TRUNCATE TABLE raw_events;` pour vider la table
2. Réexécution du script SQL corrigé
3. Vérification : 500 événements insérés avec JSON valide

**Fichiers modifiés** :
- `docs/schema.sql` (script de génération de données)

---

## 6. Let's Encrypt Timeout (Firewall)

### 🔴 Problème

```
Failed to register ACME account: Get "https://acme-v02.api.letsencrypt.org/directory": 
dial tcp 172.65.32.248:443: i/o timeout
```

Le ClusterIssuer cert-manager restait bloqué en `READY=False`. Impossible d'obtenir un certificat SSL valide.

### 🔍 Analyse

Le cluster Kubernetes sur Oracle Cloud a des **règles de sécurité réseau (Security Lists)** qui bloquent les connexions sortantes HTTPS vers Internet par défaut.

cert-manager ne peut pas contacter les serveurs ACME de Let's Encrypt pour valider le domaine et obtenir un certificat.

### ✅ Solution

**Solution choisie** : Certificat auto-signé

Création d'un ClusterIssuer et Certificate **auto-signés** :

```yaml
# selfsigned-cert.yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned-issuer
spec:
  selfSigned: {}
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: xdr-tls-selfsigned
  namespace: xdr-platform
spec:
  secretName: xdr-tls-secret
  duration: 2160h # 90 jours
  dnsNames:
    - xdr-platform.duckdns.org
  issuerRef:
    name: selfsigned-issuer
    kind: ClusterIssuer
```

**Modifications Ingress** :
Suppression de l'annotation qui causait la création automatique de certificat Let's Encrypt :
```yaml
# Commenté/supprimé :
# cert-manager.io/cluster-issuer: "letsencrypt-prod"
```

**Ordre de déploiement critique** :
1. Créer le certificat auto-signé
2. Attendre `kubectl wait --for=condition=Ready certificate/...`
3. Déployer l'Ingress

**Résultat** :
- HTTPS fonctionnel avec certificat auto-signé
- Avertissement navigateur attendu ("Non sécurisé") mais connexion chiffrée
- Acceptable pour environnement de test/portfolio

**Note pour production** :
En environnement professionnel, il faudrait ouvrir les règles de sécurité réseau ou utiliser un reverse proxy externe.

**Fichiers modifiés** :
- `kubernetes/selfsigned-cert.yaml` (nouveau)
- `kubernetes/ingress-https.yaml` (annotation supprimée)

**Commit** :
```bash
git add kubernetes/selfsigned-cert.yaml kubernetes/ingress-https.yaml
git commit -m "fix(k8s): use self-signed certificate due to Oracle Cloud firewall"
```

---

## 7. Manque de RAM (2 GB Total)

### 🔴 Problème

Instances `VM.Standard.E3.Flex` du Free Tier ont seulement **1 GB RAM par node** (2 GB total). Kafka et Zookeeper entraient régulièrement en `CrashLoopBackOff`.

### 🔍 Analyse

Les limites de ressources dans les manifests Kubernetes étaient trop élevées pour le matériel disponible. Par exemple, TimescaleDB demandait initialement 1 GB RAM = 50% d'un node.

### ✅ Solution

**1. Réduction drastique des ressources** :

```yaml
resources:
  requests:
    memory: "128Mi"  # Au lieu de 1Gi
    cpu: "50m"       # Au lieu de 500m
  limits:
    memory: "256Mi"  # Au lieu de 2Gi
    cpu: "200m"      # Au lieu de 1000m
```

**2. Désactivation de Kafka/Zookeeper** :
- Non critiques pour la démo
- Service d'ingestion écrit directement en base
- Simplifie l'architecture

**3. Priorisation des services critiques** :
- TimescaleDB : 256Mi max
- API Gateway : 128Mi max
- Frontend : 64Mi max

**Fichiers modifiés** : Tous les deployments Kubernetes

**Commit** :
```bash
git add kubernetes/*.yaml
git commit -m "perf(k8s): reduce resource limits for Oracle Free Tier (2GB RAM)"
```

---

## 8. Conflit de Certificats

### 🔴 Problème

Deux objets Certificate Kubernetes se créaient simultanément avec le même `secretName` :
- `xdr-tls-secret` (créé automatiquement par l'Ingress)
- `xdr-tls-selfsigned` (créé manuellement)

Les deux restaient en `READY=False` indéfiniment.

### 🔍 Analyse

L'annotation `cert-manager.io/cluster-issuer: "letsencrypt-prod"` dans l'Ingress provoquait la création automatique d'un Certificate par cert-manager. Ce Certificate entrait en conflit avec celui créé manuellement.

### ✅ Solution

**1. Suppression de l'annotation** dans l'Ingress :
```yaml
# annotations:
#   cert-manager.io/cluster-issuer: "letsencrypt-prod"  # ← Commenté
```

**2. Ordre de déploiement strict** :
```bash
# Supprimer tous les objets Certificate et Secrets
kubectl delete certificate --all -n xdr-platform
kubectl delete secret xdr-tls-secret -n xdr-platform
kubectl delete ingress xdr-ingress-https -n xdr-platform

# Créer d'abord le certificat auto-signé
kubectl apply -f selfsigned-cert.yaml

# Attendre qu'il soit Ready
kubectl wait --for=condition=Ready certificate/xdr-tls-selfsigned -n xdr-platform

# Puis créer l'Ingress
kubectl apply -f ingress-https.yaml
```

**Résultat** : Un seul Certificate, passe à `READY=True` en 10 secondes.

---

## 💡 Compétences Démontrées

- ✅ Analyse de logs : `kubectl logs`, `describe`, `get events`
- ✅ Compréhension des erreurs : PostgreSQL, Docker, Kubernetes, NGINX
- ✅ Recherche de solutions : Documentation officielle, GitHub Issues
- ✅ Tests itératifs : Validation après chaque modification
- ✅ Pragmatisme : Choix de contournements quand nécessaire
- ✅ Documentation : Prise de notes des erreurs et solutions
- ✅ Persévérance : Résolution de 8+ problèmes majeurs
- ✅ Adaptabilité : Ajustement de l'architecture face aux contraintes

---

## 📚 Ressources Utiles

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [TimescaleDB Documentation](https://docs.timescale.com/)
- [PostgreSQL Arrays](https://www.postgresql.org/docs/current/arrays.html)
- [Oracle Cloud Security Lists](https://docs.oracle.com/en-us/iaas/Content/Network/Concepts/securitylists.htm)
- [cert-manager Documentation](https://cert-manager.io/docs/)
