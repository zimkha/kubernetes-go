# ------------------------------
# Étape 1 : build du binaire Go
# ------------------------------
FROM golang:1.24 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o webhook main.go

# ------------------------------
# Étape 2 : image finale légère
# ------------------------------
FROM alpine:3.20

WORKDIR /root/

# Copie le binaire depuis le builder
COPY --from=builder /app/webhook .

# Copie les certificats (si tu veux embarquer les fichiers TLS)
COPY server.crt server.key /tls/

EXPOSE 8443
CMD ["./webhook"]
