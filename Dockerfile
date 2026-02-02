# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app

# Copy module files first for better layer caching
COPY go.mod ./
COPY go.sum* ./
RUN go mod download

COPY . .
# Build the server binary (static, no CGO)
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /server server_main.go

# Runtime stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates words-en
# Run as non-root for Kubernetes security
RUN adduser -D -g "" appuser
# Create /data, copy dictionary (from Alpine words-en) and ngram file while still root
RUN mkdir -p /data && chown appuser:appuser /data
RUN cp /usr/share/dict/american-english /data/dictionary.txt
COPY --from=builder /app/tetragrams-en-us.txt /data/tetragrams-en-us.txt
RUN chown appuser:appuser /data/dictionary.txt /data/tetragrams-en-us.txt
USER appuser

WORKDIR /app
COPY --from=builder /server .

EXPOSE 8080

# Dictionary and ngram file are baked in (Alpine words-en + tetragrams)
ENTRYPOINT ["/app/server"]
CMD ["--dictionary=/data/dictionary.txt", "--ngram-frequency-file=/data/tetragrams-en-us.txt"]
