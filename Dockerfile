# Stage 1: Builder
FROM golang:1.24-alpine AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/probo-api

# Stage 2: Final runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/main .
ARG DOCKER_CONFIG_PATH
RUN mkdir -p /app/config
COPY ${DOCKER_CONFIG_PATH} /app/config/docker.yaml
# COPY --from=builder /app/docker-compose.yml .
# COPY --from=builder /app/Dockerfile .

EXPOSE 8000

CMD ["./main"]