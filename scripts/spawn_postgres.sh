#!/bin/bash
set -e  
CONTAINER_NAME="chedul-db"
POSTGRES_PASSWORD="chedul"
POSTGRES_USER="chedul"
POSTGRES_DB="chedul-db"
PORT="5432"
POSTGRES_VERSION="17"

cleanup() {
    if [ $? -ne 0 ]; then
        echo "❌ El script falló. Limpiando..."
        docker rm -f $CONTAINER_NAME 2>/dev/null || true
    fi
}
trap cleanup EXIT

if ! docker info >/dev/null 2>&1; then
    echo "❌ Docker no está ejecutándose. Por favor inicia Docker e intenta nuevamente."
    exit 1
fi

if docker ps -a --format "table {{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
    echo "⚠️  El contenedor '$CONTAINER_NAME' ya existe."
    read -p "¿Quieres eliminarlo y crear uno nuevo? (s/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Ss]$ ]]; then
        echo "🗑️  Eliminando contenedor existente..."
        docker rm -f $CONTAINER_NAME
    else
        echo "❌ Cancelado. Por favor usa un nombre de contenedor diferente o elimina el contenedor existente."
        exit 1
    fi
fi

if netstat -tuln 2>/dev/null | grep -q ":${PORT} " || ss -tuln 2>/dev/null | grep -q ":${PORT} "; then
    echo "⚠️  El puerto $PORT ya está en uso. Por favor elige un puerto diferente o detén el servicio que lo está usando."
    exit 1
fi

echo "📥 Descargando imagen de PostgreSQL 17..."
docker pull postgres:$POSTGRES_VERSION

echo "🚀 Creando e iniciando contenedor PostgreSQL..."
docker run -d \
    --name $CONTAINER_NAME \
    -p 5432:5432 \
    -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD \
    -e POSTGRES_USER=$POSTGRES_USER \
    -e POSTGRES_DB=$POSTGRES_DB \
    postgres:$POSTGRES_VERSION

echo "⏳ Esperando a que PostgreSQL esté listo..."
RETRY_COUNT=0
MAX_RETRIES=30
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if docker exec $CONTAINER_NAME pg_isready -U $POSTGRES_USER -d $POSTGRES_DB >/dev/null 2>&1; then
        echo "✅ ¡PostgreSQL está listo!"
        break
    fi
    
    echo "   Verificando... (intento $((RETRY_COUNT + 1))/$MAX_RETRIES)"
    sleep 2
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo "❌ PostgreSQL falló al iniciar dentro del período de tiempo límite."
    echo "📋 Registros del contenedor:"
    docker logs $CONTAINER_NAME
    exit 1
fi

echo "🔍 Verificando estado del contenedor..."
CONTAINER_STATUS=$(docker inspect --format='{{.State.Status}}' $CONTAINER_NAME)
if [ "$CONTAINER_STATUS" != "running" ]; then
    echo "❌ El contenedor no está ejecutándose. Estado: $CONTAINER_STATUS"
    exit 1
fi

echo "🔌 Probando conexión a la base de datos..."
docker exec $CONTAINER_NAME psql -U $POSTGRES_USER -d $POSTGRES_DB -c "SELECT version();" | head -3

echo ""
echo "📊 Información del Contenedor:"
echo "   Nombre: $CONTAINER_NAME"
echo "   Estado: $(docker inspect --format='{{.State.Status}}' $CONTAINER_NAME)"
echo "   Puerto: $PORT"
echo "   Base de datos: $POSTGRES_DB"
echo "   Usuario: $POSTGRES_USER"
echo "   Contraseña: $POSTGRES_PASSWORD"
echo ""
echo "🔗 Comandos de Conexión:"
echo "   Docker exec: docker exec -it $CONTAINER_NAME psql -U $POSTGRES_USER -d $POSTGRES_DB"
echo "   Externa: psql -h localhost -p $PORT -U $POSTGRES_USER -d $POSTGRES_DB"
echo ""
echo "🛑 Para detener el contenedor: docker stop $CONTAINER_NAME"
echo "🗑️  Para eliminar el contenedor: docker rm -f $CONTAINER_NAME"
