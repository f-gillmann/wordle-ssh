# --- Build ---
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o wordle-ssh .

# --- Run ---
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/wordle-ssh .

RUN mkdir -p .ssh

EXPOSE 23234

ENV WORDLE_SSH_HOST=0.0.0.0
ENV WORDLE_SSH_PORT=23234
ENV WORDLE_SSH_HOST_KEY_PATH=.ssh/id_ed25519

CMD ["./wordle-ssh"]
