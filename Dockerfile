# Build multi-stage: compila em uma imagem completa do Go, roda em uma
# imagem mínima. Mantém a imagem final pequena e sem toolchain de build.

FROM golang:1.26-alpine AS builder
WORKDIR /src

# Cacheia o download de módulos separado da cópia do código-fonte.
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/gestorbuy-api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/gestorbuy-api /usr/local/bin/gestorbuy-api

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/gestorbuy-api"]
