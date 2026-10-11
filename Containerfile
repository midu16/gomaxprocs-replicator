# Use explicit Go 1.24
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY main.go .
# Compile a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o replicator main.go

FROM scratch
COPY --from=builder /app/replicator /replicator
ENTRYPOINT ["/replicator"]
