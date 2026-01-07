#!/bin/bash

# Script de déploiement Kubernetes pour XDR Platform
# Usage: ./deploy.sh

set -e

echo "🚀 Déploiement de XDR Platform sur Kubernetes..."

# Couleurs pour l'affichage
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Vérifier que kubectl est installé
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}❌ kubectl n'est pas installé${NC}"
    exit 1
fi

echo -e "${GREEN}✅ kubectl trouvé${NC}"

# Vérifier la connexion au cluster
echo -e "${YELLOW}📡 Vérification de la connexion au cluster...${NC}"
if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}❌ Impossible de se connecter au cluster Kubernetes${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Connecté au cluster${NC}"

# Créer le namespace
echo -e "${YELLOW}📦 Création du namespace...${NC}"
kubectl apply -f k8s/00-namespace.yaml

# Créer les ConfigMaps et Secrets
echo -e "${YELLOW}🔐 Création des ConfigMaps et Secrets...${NC}"
kubectl apply -f k8s/01-configmap.yaml
kubectl apply -f k8s/02-secrets.yaml

# Créer les PVCs
echo -e "${YELLOW}💾 Création des PersistentVolumeClaims...${NC}"
kubectl apply -f k8s/03-pvcs.yaml

# Attendre que les PVCs soient bound
echo -e "${YELLOW}⏳ Attente du provisionnement des volumes...${NC}"
kubectl wait --for=condition=Bound pvc --all -n xdr-platform --timeout=300s

# Déployer l'infrastructure (DB, Kafka, Redis)
echo -e "${YELLOW}🗄️  Déploiement de l'infrastructure...${NC}"
kubectl apply -f k8s/10-timescaledb.yaml
kubectl apply -f k8s/11-redis.yaml
kubectl apply -f k8s/12-zookeeper.yaml
kubectl apply -f k8s/13-kafka.yaml

# Attendre que l'infrastructure soit prête
echo -e "${YELLOW}⏳ Attente du démarrage de l'infrastructure...${NC}"
kubectl wait --for=condition=available --timeout=300s deployment/timescaledb -n xdr-platform
kubectl wait --for=condition=available --timeout=300s deployment/redis -n xdr-platform
kubectl wait --for=condition=available --timeout=300s deployment/zookeeper -n xdr-platform
kubectl wait --for=condition=available --timeout=300s deployment/kafka -n xdr-platform

# Déployer les services applicatifs
echo -e "${YELLOW}🚀 Déploiement des services applicatifs...${NC}"
kubectl apply -f k8s/20-agent.yaml
kubectl apply -f k8s/21-ingestion.yaml
kubectl apply -f k8s/22-api-gateway.yaml
kubectl apply -f k8s/23-frontend.yaml

# Attendre que les services soient prêts
echo -e "${YELLOW}⏳ Attente du démarrage des services...${NC}"
kubectl wait --for=condition=available --timeout=300s deployment/xdr-agent -n xdr-platform
kubectl wait --for=condition=available --timeout=300s deployment/xdr-ingestion -n xdr-platform
kubectl wait --for=condition=available --timeout=300s deployment/xdr-api-gateway -n xdr-platform
kubectl wait --for=condition=available --timeout=300s deployment/xdr-frontend -n xdr-platform

# Déployer l'Ingress (optionnel)
if [ -f "k8s/30-ingress.yaml" ]; then
    echo -e "${YELLOW}🌐 Déploiement de l'Ingress...${NC}"
    kubectl apply -f k8s/30-ingress.yaml
fi

echo ""
echo -e "${GREEN}✅ Déploiement terminé avec succès !${NC}"
echo ""
echo -e "${YELLOW}📊 État des déploiements:${NC}"
kubectl get deployments -n xdr-platform

echo ""
echo -e "${YELLOW}🌐 Services:${NC}"
kubectl get services -n xdr-platform

echo ""
echo -e "${YELLOW}📦 Pods:${NC}"
kubectl get pods -n xdr-platform

echo ""
echo -e "${GREEN}🎉 XDR Platform est maintenant déployée !${NC}"
echo ""

# Afficher l'URL d'accès
FRONTEND_IP=$(kubectl get svc frontend-service -n xdr-platform -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "pending")
if [ "$FRONTEND_IP" != "pending" ]; then
    echo -e "${GREEN}🌍 URL d'accès: http://${FRONTEND_IP}${NC}"
else
    echo -e "${YELLOW}⏳ L'IP publique est en cours de provisionnement...${NC}"
    echo -e "${YELLOW}   Utilisez: kubectl get svc frontend-service -n xdr-platform${NC}"
fi

echo ""
echo -e "${YELLOW}📝 Commandes utiles:${NC}"
echo "  - Voir les logs:        kubectl logs -f deployment/xdr-api-gateway -n xdr-platform"
echo "  - Voir tous les pods:   kubectl get pods -n xdr-platform"
echo "  - Redémarrer un pod:    kubectl rollout restart deployment/xdr-frontend -n xdr-platform"
echo "  - Supprimer tout:       kubectl delete namespace xdr-platform"
