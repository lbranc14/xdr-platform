# Dockerfile pour l'Agent XDR
FROM golang:1.21-alpine AS builder

# Installer les dépendances système
RUN apk add --no-cache git

# Définir le répertoire de travail
WORKDIR /app

# Copier les fichiers go.mod et go.sum
COPY go.mod go.sum ./

# Télécharger les dépendances
RUN go mod download

# Copier tout le code source
COPY . .

# Compiler l'agent
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o xdr-agent .

# Image finale légère
FROM alpine:latest

# Installer ca-certificates pour HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copier le binaire compilé
COPY --from=builder /app/xdr-agent .

# Variables d'environnement par défaut
ENV AGENT_ID=""
ENV KAFKA_BROKERS="kafka:9092"
ENV KAFKA_TOPIC="raw-events"
ENV COLLECTION_INTERVAL="30s"

# Lancer l'agent
CMD ["./xdr-agent"]
