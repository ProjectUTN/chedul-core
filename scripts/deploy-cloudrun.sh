#!/usr/bin/env bash
# Despliega la API en Google Cloud Run desde el codigo (Cloud Build arma la
# imagen con el Dockerfile). La primera vez pide JWT_SECRET y DATABASE_URL;
# las siguientes solo actualiza el codigo y conserva las variables.
#
# Uso:  ./scripts/deploy-cloudrun.sh
# Variables opcionales: GCP_PROJECT, REGION (default southamerica-east1),
# SERVICE (default chedul-api).
set -euo pipefail

REGION="${REGION:-southamerica-east1}"
SERVICE="${SERVICE:-chedul-api}"
PROJECT="${GCP_PROJECT:-$(gcloud config get-value project 2>/dev/null)}"

if [[ -z "$PROJECT" ]]; then
  echo "Elegí un proyecto: gcloud config set project TU_PROYECTO" >&2
  exit 1
fi

args=(
  run deploy "$SERVICE"
  --project "$PROJECT"
  --region "$REGION"
  --source .
  --allow-unauthenticated
  --memory 256Mi
  --cpu 1
  --min-instances 0
  --max-instances 2
  --cpu-boost
)

# Primera vez: el servicio todavia no existe y hay que pasarle la config
if ! gcloud run services describe "$SERVICE" --project "$PROJECT" --region "$REGION" >/dev/null 2>&1; then
  read -rp "DATABASE_URL (la connection string de Neon): " database_url
  jwt_secret="$(openssl rand -hex 32)"
  args+=(--set-env-vars "^##^DATABASE_URL=${database_url}##JWT_SECRET=${jwt_secret}##UPLOADS_DISABLED=true")
  echo "Se generó un JWT_SECRET nuevo y se guardó en el servicio."
fi

gcloud "${args[@]}"

echo
echo "Listo. URL de la API:"
gcloud run services describe "$SERVICE" --project "$PROJECT" --region "$REGION" --format 'value(status.url)'
