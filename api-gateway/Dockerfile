# Dockerfile pour l'API Gateway
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copier les dépendances
COPY go.mod go.sum ./
RUN go mod download

# Copier le code
COPY . .

# Compiler
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-gateway .

# Image finale
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/api-gateway .

# Variables d'environnement
ENV DATABASE_HOST="timescaledb"
ENV DATABASE_PORT="5432"
ENV DATABASE_NAME="xdr_events"
ENV DATABASE_USER="xdr_admin"
ENV DATABASE_PASSWORD="xdr_secure_password_2024"
ENV API_PORT="8000"

# Exposer le port
EXPOSE 8000

# Lancer l'API
CMD ["./api-gateway"]
