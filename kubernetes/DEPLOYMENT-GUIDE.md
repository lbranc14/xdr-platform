# 🚀 Guide de Déploiement sur Oracle Cloud

Ce guide vous accompagne pour déployer votre plateforme XDR sur Oracle Cloud Kubernetes (OKE) avec le Free Tier.

## 📋 Prérequis

- Compte Oracle Cloud (gratuit)
- `kubectl` installé localement
- `oci` CLI installé (Oracle Cloud CLI)
- Docker Hub ou un registry Docker

---

## 🎯 Étape 1 : Créer un compte Oracle Cloud

1. Allez sur https://www.oracle.com/cloud/free/
2. Créez un compte gratuit (Always Free)
3. Validez votre email et configurez votre compte

**Ce qui est gratuit à vie:**
- 2 VMs AMD (1/8 OCPU, 1 GB RAM chacune)
- 4 VMs ARM Ampere A1 (24 GB RAM, 4 OCPUs au total)
- 200 GB Block Volume
- 20 GB Object Storage
- 1 flexible Load Balancer

---

## 🏗️ Étape 2 : Créer un cluster Kubernetes (OKE)

### Via l'interface web (recommandé):

1. **Connectez-vous** à Oracle Cloud Console
2. Menu **≡** → **Developer Services** → **Kubernetes Clusters (OKE)**
3. Cliquez sur **Create Cluster**
4. Choisissez **Quick Create**
5. Configuration:
   - **Name**: `xdr-cluster`
   - **Kubernetes Version**: Latest
   - **Node Pool**:
     - **Shape**: `VM.Standard.A1.Flex` (ARM - Always Free)
     - **OCPUs**: 4 (ou 2 si vous voulez garder pour autre chose)
     - **Memory**: 24 GB
     - **Number of nodes**: 2
   - **Network**: Utiliser les valeurs par défaut (VCN auto-créé)
6. Cliquez sur **Create Cluster**
7. **Attendre ~10 minutes** que le cluster soit créé

### Vérifier le cluster:

```bash
# Le cluster doit être en état "Active"
```

---

## 🔐 Étape 3 : Configurer kubectl

1. Dans OKE Console, cliquez sur votre cluster `xdr-cluster`
2. Cliquez sur **Access Cluster**
3. Suivez les instructions affichées:

```bash
# Exemple (vos valeurs seront différentes):
oci ce cluster create-kubeconfig \
  --cluster-id ocid1.cluster.oc1... \
  --file $HOME/.kube/config \
  --region us-phoenix-1
```

4. Vérifiez la connexion:

```bash
kubectl get nodes
```

Vous devriez voir vos 2 nodes ARM.

---

## 🐳 Étape 4 : Pusher vos images Docker

### Option A : Docker Hub (gratuit)

```bash
# 1. Se connecter à Docker Hub
docker login

# 2. Tag les images
docker tag xdr-platform-agent:latest votre-username/xdr-agent:latest
docker tag xdr-platform-ingestion:latest votre-username/xdr-ingestion:latest
docker tag xdr-platform-api-gateway:latest votre-username/xdr-api-gateway:latest
docker tag xdr-platform-frontend:latest votre-username/xdr-frontend:latest

# 3. Push les images
docker push votre-username/xdr-agent:latest
docker push votre-username/xdr-ingestion:latest
docker push votre-username/xdr-api-gateway:latest
docker push votre-username/xdr-frontend:latest
```

### Option B : Oracle Container Registry (OCIR)

```bash
# 1. Se connecter à OCIR
docker login <region-key>.ocir.io
# Username: <tenancy-namespace>/<votre-username>
# Password: <auth-token>

# 2. Tag les images
docker tag xdr-platform-agent:latest <region>.ocir.io/<namespace>/xdr-agent:latest
# etc...

# 3. Push les images
docker push <region>.ocir.io/<namespace>/xdr-agent:latest
```

---

## 📝 Étape 5 : Mettre à jour les manifests

Dans chaque fichier de déploiement (`k8s/20-*.yaml`), remplacez:

```yaml
image: your-docker-registry/xdr-agent:latest
```

Par:

```yaml
image: votre-username/xdr-agent:latest  # Docker Hub
# OU
image: <region>.ocir.io/<namespace>/xdr-agent:latest  # OCIR
```

Fichiers à modifier:
- `k8s/20-agent.yaml`
- `k8s/21-ingestion.yaml`
- `k8s/22-api-gateway.yaml`
- `k8s/23-frontend.yaml`

---

## 🚀 Étape 6 : Déployer sur Kubernetes

```bash
# Donner les droits d'exécution au script
chmod +x k8s/deploy.sh

# Lancer le déploiement
./k8s/deploy.sh
```

Le script va:
1. ✅ Créer le namespace
2. ✅ Créer les ConfigMaps et Secrets
3. ✅ Provisionner les volumes persistants
4. ✅ Déployer TimescaleDB, Redis, Kafka
5. ✅ Déployer vos services applicatifs
6. ✅ Exposer le frontend via LoadBalancer

---

## 🌐 Étape 7 : Accéder à votre plateforme

```bash
# Récupérer l'IP publique du LoadBalancer
kubectl get svc frontend-service -n xdr-platform

# Exemple de sortie:
# NAME               TYPE           EXTERNAL-IP       PORT(S)
# frontend-service   LoadBalancer   140.238.123.45    80:31234/TCP
```

Votre plateforme est accessible sur: **http://140.238.123.45**

---

## 🔒 Étape 8 (Optionnel) : Configurer HTTPS

### Installer cert-manager:

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
```

### Créer un ClusterIssuer Let's Encrypt:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: votre-email@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF
```

### Installer NGINX Ingress Controller:

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.9.4/deploy/static/provider/cloud/deploy.yaml
```

### Configurer votre domaine:

1. Achetez un domaine (Namecheap, GoDaddy, etc.) ou utilisez un domaine gratuit (Freenom)
2. Créez un enregistrement A pointant vers l'IP du LoadBalancer
3. Modifiez `k8s/30-ingress.yaml` et remplacez `your-domain.com`
4. Déployez l'Ingress:

```bash
kubectl apply -f k8s/30-ingress.yaml
```

Votre plateforme sera accessible sur: **https://votre-domaine.com** 🔒

---

## 📊 Surveillance et Monitoring

### Voir les logs en temps réel:

```bash
# Logs de l'API Gateway
kubectl logs -f deployment/xdr-api-gateway -n xdr-platform

# Logs de l'ingestion
kubectl logs -f deployment/xdr-ingestion -n xdr-platform

# Tous les logs
kubectl logs -f -l app=xdr-api-gateway -n xdr-platform --all-containers
```

### Dashboard Kubernetes:

```bash
# Installer le dashboard
kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.7.0/aio/deploy/recommended.yaml

# Créer un token d'accès
kubectl -n kubernetes-dashboard create token admin-user

# Port-forward pour accéder
kubectl port-forward -n kubernetes-dashboard service/kubernetes-dashboard 8443:443
```

Accès: https://localhost:8443

---

## 🔧 Commandes utiles

```bash
# Voir tous les pods
kubectl get pods -n xdr-platform

# Voir les services
kubectl get svc -n xdr-platform

# Redémarrer un déploiement
kubectl rollout restart deployment/xdr-frontend -n xdr-platform

# Scaler un déploiement
kubectl scale deployment/xdr-agent --replicas=5 -n xdr-platform

# Entrer dans un pod
kubectl exec -it <pod-name> -n xdr-platform -- /bin/sh

# Voir les événements
kubectl get events -n xdr-platform --sort-by='.lastTimestamp'

# Supprimer tout
kubectl delete namespace xdr-platform
```

---

## 💰 Coûts estimés

Avec le **Always Free Tier**:
- ✅ **0€/mois** tant que vous restez dans les limites gratuites
- ✅ 4 OCPUs ARM + 24 GB RAM (amplement suffisant)
- ✅ 200 GB de stockage

**Au-delà du Free Tier** (si vous scalez):
- ~50-100€/mois pour un cluster plus important
- Load Balancer: ~15€/mois
- Block Storage: ~0.05€/GB/mois

---

## 🆘 Dépannage

### Les pods ne démarrent pas:

```bash
# Voir les détails du pod
kubectl describe pod <pod-name> -n xdr-platform

# Voir les événements
kubectl get events -n xdr-platform
```

### Problème de pull d'image:

Vérifiez que les images sont publiques sur Docker Hub ou créez un ImagePullSecret pour OCIR.

### Manque de ressources:

```bash
# Voir l'utilisation des ressources
kubectl top nodes
kubectl top pods -n xdr-platform
```

Réduisez les `replicas` ou les `resources.limits` dans les manifests.

---

## 🎉 Félicitations !

Votre plateforme XDR est maintenant **déployée en production** sur Oracle Cloud avec Kubernetes ! 🚀

---

## 📚 Ressources

- [Oracle Cloud Free Tier](https://www.oracle.com/cloud/free/)
- [OKE Documentation](https://docs.oracle.com/en-us/iaas/Content/ContEng/home.htm)
- [Kubernetes Documentation](https://kubernetes.io/docs/home/)
- [cert-manager Documentation](https://cert-manager.io/docs/)
