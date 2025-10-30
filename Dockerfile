# --- Build ---
FROM golang:1.23-alpine AS builder

RUN apk --no-cache add gcc musl-dev sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o wordle-ssh ./cmd/main.go

# --- Run ---
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /root/

COPY --from=builder /app/wordle-ssh .

# Create directories for SSH keys and database
RUN mkdir -p .ssh /data

EXPOSE 23234

ENV WORDLE_SSH_HOST=0.0.0.0
ENV WORDLE_SSH_PORT=23234
ENV WORDLE_SSH_HOST_KEY_PATH=.ssh/id_ed25519
ENV WORDLE_SSH_DB_PATH=/data/wordle-stats.db

CMD ["./wordle-ssh"]
