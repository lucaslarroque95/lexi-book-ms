FROM golang:1.26 AS builder

WORKDIR /app

COPY src/go.mod src/go.sum ./
RUN go mod download

COPY src/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /out/server ./server
COPY src/keys ./keys

RUN adduser -D appuser
USER appuser

EXPOSE 8080

CMD ["./server"]
