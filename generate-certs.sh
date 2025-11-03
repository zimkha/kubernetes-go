#!/bin/bash
set -e

# Variables
SERVICE="label-validator"
NAMESPACE="kube-system"
SECRET_NAME="label-validator-tls"

# 1️⃣ Crée une clé privée pour la CA
openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key -subj "/CN=admission_ca" -days 365 -out ca.crt

# 2️⃣ Crée la clé et la CSR (certificate signing request) pour le service webhook
openssl genrsa -out server.key 2048
openssl req -new -key server.key -subj "/CN=${SERVICE}.${NAMESPACE}.svc" -out server.csr

# 3️⃣ Signe le certificat serveur avec la CA
cat > server-ext.cnf <<EOF
[ v3_ext ]
subjectAltName = @alt_names

[ alt_names ]
DNS.1 = ${SERVICE}
DNS.2 = ${SERVICE}.${NAMESPACE}
DNS.3 = ${SERVICE}.${NAMESPACE}.svc
EOF

openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out server.crt -days 365 -extensions v3_ext -extfile server-ext.cnf

# 4️⃣ Crée le secret Kubernetes avec les fichiers TLS
kubectl create secret tls ${SECRET_NAME} \
  --cert=server.crt \
  --key=server.key \
  -n ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -

# 5️⃣ Affiche le CA en base64 (à mettre dans ton webhook.yaml)
echo ""
echo "🔑 Copie ce bloc et colle-le dans ton webhook.yaml sous 'caBundle':"
echo ""
cat ca.crt | base64 | tr -d '\n'
echo ""
