# --- Build ---
FROM golang:1.25-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -o wordle-ssh ./cmd/main.go

# --- Run ---
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libsqlite3-0 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /root/

COPY --from=builder /app/wordle-ssh .

# Create directories for SSH keys and database
RUN mkdir -p .ssh /data

EXPOSE 23234

ENV TERM=xterm-256color
ENV WORDLE_SSH_HOST="0.0.0.0"
ENV WORDLE_SSH_PORT="23234"
ENV WORDLE_SSH_MOTD="Welcome to Wordle SSH!"
ENV WORDLE_SSH_HOST_KEY_PATH=".ssh/id_ed25519"
ENV WORDLE_SSH_DB_PATH="/data/wordle-stats.db"

CMD ["./wordle-ssh"]
