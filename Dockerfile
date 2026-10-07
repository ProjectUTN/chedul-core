FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o main ./cmd/api

# Directorio para los archivos subidos. Solo sirve con un disco persistente
# montado aca; en Cloud Run se usa UPLOADS_DISABLED=true y los aportes son links.
RUN mkdir -p /out/uploads

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /app/configuration /configuration
COPY --from=builder /app/main /main
COPY --from=builder /out/uploads /uploads

# Corre como root: los volumenes de Railway o de un VPS se montan con dueño
# root y un usuario sin privilegios no podria escribir los archivos subidos.
# La imagen es "scratch" (no tiene shell ni nada mas que el binario).
# Cloud Run y la mayoria de los hostings definen PORT y la API lo respeta.
ENV CONFIG_DIR=configuration/ \
    SERVER_ENV=production \
    UPLOADS_DIR=/uploads

EXPOSE 8080

ENTRYPOINT ["/main"]
