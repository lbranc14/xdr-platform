# Dockerfile pour le Service d'Ingestion
FROM golang:1.21-alpine AS builder

# Installer les dépendances système
RUN apk add --no-cache git

WORKDIR /app

# Copier les fichiers de dépendances
COPY go.mod go.sum ./
RUN go mod download

# Copier le code source
COPY . .

# Compiler le service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ingestion-service .

# Image finale
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copier le binaire
COPY --from=builder /app/ingestion-service .

# Variables d'environnement
ENV DATABASE_HOST="timescaledb"
ENV DATABASE_PORT="5432"
ENV DATABASE_NAME="xdr_events"
ENV DATABASE_USER="xdr_admin"
ENV DATABASE_PASSWORD="xdr_secure_password_2024"
ENV KAFKA_BROKERS="kafka:9092"
ENV KAFKA_TOPIC_RAW_EVENTS="raw-events"
ENV KAFKA_GROUP_ID="xdr-ingestion-service"

# Lancer le service
CMD ["./ingestion-service"]
