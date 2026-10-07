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

# Directorio para los archivos subidos. En produccion hay que montar un volumen aca.
RUN mkdir -p /out/uploads

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /app/configuration /configuration
COPY --from=builder /app/main /main
COPY --from=builder /out/uploads /uploads

# Corre como root: los volumenes de Fly.io y Railway se montan con dueño root
# y un usuario sin privilegios no podria escribir los archivos subidos.
# La imagen es "scratch" (no tiene shell ni nada mas que el binario).
ENV CONFIG_DIR=configuration/ \
    SERVER_ENV=production \
    UPLOADS_DIR=/uploads

EXPOSE 8080

ENTRYPOINT ["/main"]
