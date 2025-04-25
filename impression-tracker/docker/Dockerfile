# 1) Builder stage: compile l'app sous Alpine
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Dépendances système pour Go & git
RUN apk add --no-cache git

# Modules Go
COPY go.mod go.sum ./
RUN go mod download

# Code source
COPY . .

# Build statique
RUN CGO_ENABLED=0 GOOS=linux go build -o impression-tracker ./cmd/main.go

# 2) Runtime stage: Alpine minimal avec .env support
FROM alpine:3.21

WORKDIR /app

# Pour que Go puisse faire des appels TLS si besoin
RUN apk add --no-cache ca-certificates

# Copie du binaire
COPY --from=builder /app/impression-tracker .

# Expose le port gRPC
EXPOSE 50052

# Lancement
ENTRYPOINT ["./impression-tracker"]
 